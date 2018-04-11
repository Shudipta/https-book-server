/*
Copyright 2018 The Soter Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package fake

import (
	v1alpha1 "github.com/soter/scanner/apis/scanner/v1alpha1"
	schema "k8s.io/apimachinery/pkg/runtime/schema"
	testing "k8s.io/client-go/testing"
)

// FakeImageReviews implements ImageReviewInterface
type FakeImageReviews struct {
	Fake *FakeScannerV1alpha1
	ns   string
}

var imagereviewsResource = schema.GroupVersionResource{Group: "scanner.soter.cloud", Version: "v1alpha1", Resource: "imagereviews"}

var imagereviewsKind = schema.GroupVersionKind{Group: "scanner.soter.cloud", Version: "v1alpha1", Kind: "ImageReview"}

// Create takes the representation of a imageReview and creates it.  Returns the server's representation of the imageReview, and an error, if there is any.
func (c *FakeImageReviews) Create(imageReview *v1alpha1.ImageReview) (result *v1alpha1.ImageReview, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewCreateAction(imagereviewsResource, c.ns, imageReview), &v1alpha1.ImageReview{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.ImageReview), err
}