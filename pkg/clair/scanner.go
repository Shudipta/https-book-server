package clair

import (
	"encoding/base64"
	"fmt"
	"net/http"

	"github.com/coreos/clair/api/v3/clairpb"
	manifestV1 "github.com/docker/distribution/manifest/schema1"
	manifestV2 "github.com/docker/distribution/manifest/schema2"
	reg "github.com/heroku/docker-registry-client/registry"
	"github.com/pkg/errors"
	api "github.com/soter/scanner/apis/scanner/v1alpha1"
)

// This method takes <registryUrl>, <imageName>, <username>, <password> and
// returns (true, error) if any error occurred or vulnerability found otherwise return
// (false, <nil>). First, it parses the <imageName> and connect to the registry. Then,
// it gets the manifests for the <imageName>. Then for each layer in the manifests it
// sends a LayerType{} obj to scanner. Then it does a GET requests to clair and receives
// a LayerType{} obj as response.Body. Then it filters this layer obj to get the features
// and vulnerabilities. Finally, it stores them in cache against the layer name as key.
// It stores them so that it finds them in cache without calling to clair and filtering
// again if next time we need to scan same layer.
// For more information about LayerType{}, https://coreos.com/clair/docs/latest/api_v1.html
// will be helpful.
func IsVulnerable(
	clairAddress string,
	registryUrl, imageName, username, password string) ([]api.Feature, []api.Vulnerability, error) {

	// TODO: need to check for digest part
	repo, tag, _, err := parseImageName(imageName)
	if err != nil {
		return nil, nil, WithCode(err, ParseImageNameError)
	}

	hub := &reg.Registry{
		URL: registryUrl,
		Client: &http.Client{
			Transport: reg.WrapTransport(http.DefaultTransport, registryUrl, username, password),
		},
		Logf: reg.Quiet,
	}
	mx, err := hub.ManifestVx(repo, tag)
	if err != nil {
		return nil, nil, WithCode(errors.Wrapf(err, "failed to retrieve manifest for image %s", imageName), GettingManifestError)
	}

	postAncestryRequest := &clairpb.PostAncestryRequest{
		AncestryName: repo,
		Format:       "Docker",
	}
	switch manifest := mx.(type) {
	case *manifestV2.DeserializedManifest:
		layers := make([]*clairpb.PostAncestryRequest_PostLayer, len(manifest.Layers))
		for i, layer := range manifest.Layers {
			layers[i] = &clairpb.PostAncestryRequest_PostLayer{
				Hash: hashPart(manifest.Config.Digest.String()) + hashPart(layer.Digest.String()),
				Path: fmt.Sprintf("%s/%s/%s/%s", registryUrl, repo, "blobs", layer.Digest.String()),
				Headers: map[string]string{
					"Authorization": "Basic " + base64.StdEncoding.EncodeToString([]byte(username+":"+password)),
				},
			}
		}
		postAncestryRequest.Layers = layers
	case *manifestV1.SignedManifest:
		layers := make([]*clairpb.PostAncestryRequest_PostLayer, len(manifest.FSLayers))
		for i, layer := range manifest.FSLayers {
			layers[len(manifest.FSLayers)-1-i] = &clairpb.PostAncestryRequest_PostLayer{
				Hash: hashPart(layer.BlobSum.String()),
				Path: fmt.Sprintf("%s/%s/%s/%s", registryUrl, repo, "blobs", layer.BlobSum.String()),
				Headers: map[string]string{
					"Authorization": "Basic " + base64.StdEncoding.EncodeToString([]byte(username+":"+password)),
				},
			}
		}
		postAncestryRequest.Layers = layers
	default:
		return nil, nil, WithCode(errors.New("unknown manifest type"), UnknownManifestError)
	}

	layersLen := len(postAncestryRequest.Layers)
	if layersLen == 0 {
		return nil, nil, WithCode(errors.Wrapf(err, "failed to pull Layers for image %s", imageName), PullingLayersError)
	} else {
		fmt.Println("Analyzing", layersLen, "layers")
	}

	clairClient, err := clairClientSetup(clairAddress)
	if err != nil {
		return nil, nil, WithCode(errors.Wrapf(err, "failed to connect"), ConnectingClairClientError)
	}

	err = sendLayer(postAncestryRequest, clairClient)
	if err != nil {
		return nil, nil, WithCode(errors.Wrapf(err, "failed to send layers for image %s", imageName), PostAncestryError)
	}

	features, vulnerabilities, err := getLayer(repo, clairClient)
	if err != nil {
		return nil, nil, WithCode(errors.Wrapf(err, "failed to get features and vulnerabilities for image %s", imageName), GetAncestryError)
	}

	if vulnerabilities != nil {
		return features, vulnerabilities, WithCode(errors.Errorf("Image %s contains vulnerabilities", imageName), VulnerableStatus)
	}

	return features, vulnerabilities, WithCode(nil, NotVulnerableStatus)
}