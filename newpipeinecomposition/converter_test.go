package newpipeinecomposition

import (
	"testing"

	v1 "github.com/crossplane/crossplane/apis/apiextensions/v1"
	"github.com/google/go-cmp/cmp"
)

func TestSetMissingConnectionDetailFields(t *testing.T) {
	kubeconfigKey := "kubeconfig"
	fv := v1.ConnectionDetailTypeFromValue
	ffp := v1.ConnectionDetailTypeFromFieldPath
	fcsk := v1.ConnectionDetailTypeFromConnectionSecretKey
	type args struct {
		sk v1.ConnectionDetail
	}
	type want struct {
		sk v1.ConnectionDetail
	}

	cases := map[string]struct {
		reason string
		args   args
		want   want
	}{
		"ConnectionDetailMissingKeyAndName": {
			reason: "Correctly add Type and Name",
			args: args{
				sk: v1.ConnectionDetail{
					FromConnectionSecretKey: &kubeconfigKey,
				},
			},
			want: want{
				sk: v1.ConnectionDetail{
					Name:                    &kubeconfigKey,
					FromConnectionSecretKey: &kubeconfigKey,
					Type:                    &fcsk,
				},
			},
		},
		"FromValueMissingType": {
			reason: "Correctly add Type",
			args: args{
				sk: v1.ConnectionDetail{
					Name:  &kubeconfigKey,
					Value: &kubeconfigKey,
				},
			},
			want: want{
				sk: v1.ConnectionDetail{
					Name:  &kubeconfigKey,
					Value: &kubeconfigKey,
					Type:  &fv,
				},
			},
		},
		"FromFieldPathMissingType": {
			reason: "Correctly add Type",
			args: args{
				sk: v1.ConnectionDetail{
					Name:          &kubeconfigKey,
					FromFieldPath: &kubeconfigKey,
				},
			},
			want: want{
				sk: v1.ConnectionDetail{
					Name:          &kubeconfigKey,
					FromFieldPath: &kubeconfigKey,
					Type:          &ffp,
				},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			sk := SetMissingConnectionDetailFields(tc.args.sk)
			if diff := cmp.Diff(tc.want.sk, sk); diff != "" {
				t.Errorf("%s\nPopulateConnectionSecret(...): -want i, +got i:\n%s", tc.reason, diff)
			}

		})
	}
}

func TestSetTransformTypeRequiredFields(t *testing.T) {
	mult := int64(1024)
	type args struct {
		tt v1.Transform
	}
	type want struct {
		tt v1.Transform
	}

	cases := map[string]struct {
		reason string
		args   args
		want   want
	}{
		"MathMissingTYpe": {
			reason: "Correctly add Type and Name",
			args: args{
				tt: v1.Transform{
					Math: &v1.MathTransform{Multiply: &mult},
				},
			},
			want: want{
				tt: v1.Transform{
					Type: v1.TransformTypeMath,
					Math: &v1.MathTransform{Multiply: &mult, Type: v1.MathTransformTypeMultiply},
				},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			tt := SetTransformTypeRequiredFields(tc.args.tt)
			if diff := cmp.Diff(tc.want.tt, tt); diff != "" {
				t.Errorf("%s\nPopulateTransformType(...): -want i, +got i:\n%s", tc.reason, diff)
			}

		})
	}
}

func EquateErrors() cmp.Option {
	return cmp.Comparer(func(a, b error) bool {
		if a == nil || b == nil {
			return a == nil && b == nil
		}
		return a.Error() == b.Error()
	})
}
