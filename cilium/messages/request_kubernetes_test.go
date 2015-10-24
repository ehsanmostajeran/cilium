package messages

import (
	"reflect"
	"strconv"
	"strings"
	"testing"

	k8s "github.com/cilium-team/cilium/Godeps/_workspace/src/k8s.io/kubernetes/pkg/api"
)

var (
	validKubernetesSimpleRequest = `{"Type": ` + validType + `, "PowerstripProtocolVersion": ` +
		strconv.Itoa(validPPV) + `, "ClientRequest": {"Body": "` + validKubernetesRequest + `", ` +
		`"Request": ` + validRequestHeader + `, "Method": "` + validMethod + `"}}`
	validKubernetesRequest = `{\"kind\":\"Pod\",\"apiVersion\":\"v1\",\"metadata\":` +
		`{\"name\":\"redis-controller-tester\",\"namespace\":\"default\",\"creationTimestamp\":null,` +
		`\"labels\":{\"app\":\"redis-controller\",\"com.docker.compose.service\":` +
		`\"redis-controller\"}},\"spec\":{\"containers\":[{\"name\":\"redis\",\"image\":` +
		`\"redis\",\"ports\":[{\"containerPort\":6379,\"protocol\":\"TCP\"}],\"resources\":` +
		`{},\"terminationMessagePath\":\"/dev/termination-log\",\"imagePullPolicy\":` +
		`\"IfNotPresent\"}],\"restartPolicy\":\"Always\",\"dnsPolicy\":\"ClusterFirst\"},` +
		`\"status\":{}}`
	validKubernetesRequestWoutEscQuot = strings.Replace(validKubernetesRequest, `\"`, `"`, -1)
	validKubernetesObjReference       = k8s.ObjectReference{
		APIVersion:      "v1",
		Kind:            "Pod",
		ResourceVersion: "thisdoesntexistonejsonrequest",
	}
	validKubernetesBodyObj = map[string]interface{}{
		"kind":            "Pod",
		"apiVersion":      "v1",
		"resourceVersion": "thisdoesntexistonejsonrequest",
		"metadata": map[string]interface{}{
			"name":              "redis-controller-tester",
			"namespace":         "default",
			"creationTimestamp": nil,
			"labels": map[string]interface{}{
				"app": "redis-controller",
				"com.docker.compose.service": "redis-controller",
			},
		},
		"spec": map[string]interface{}{
			"containers": []interface{}{
				map[string]interface{}{
					"name":  "redis",
					"image": "redis",
					"ports": []interface{}{
						map[string]interface{}{
							"containerPort": float64(6379),
							"protocol":      "TCP",
						},
					},
					"resources":              map[string]interface{}{},
					"terminationMessagePath": "/dev/termination-log",
					"imagePullPolicy":        "IfNotPresent",
				},
			},
			// When the kubernetes PR is accepted we can remove the following
			// 2 lines.
			"volumes":            nil,
			"serviceAccountName": "",
			"restartPolicy":      "Always",
			"dnsPolicy":          "ClusterFirst",
		},
		"status": map[string]interface{}{},
	}
)

func TestKubernetesUnmarshalKubernetesObjRefClientBody(t *testing.T) {
	pr := PowerstripRequest{
		ClientRequest: ClientRequest{
			Body: validKubernetesRequestWoutEscQuot,
		},
	}
	var kor KubernetesObjRef
	if err := pr.UnmarshalKubernetesObjRefClientBody(&kor); err != nil {
		t.Fatal("invalid validKubernetesRequestWoutEscQuot:", err)
	}
	kor.ResourceVersion = "thisdoesntexistonejsonrequest"
	kor.BodyObj["resourceVersion"] = "thisdoesntexistonejsonrequest"
	korWant := KubernetesObjRef{
		ObjectReference: validKubernetesObjReference,
		BodyObj:         validKubernetesBodyObj,
	}
	if !reflect.DeepEqual(kor.ObjectReference, korWant.ObjectReference) {
		t.Errorf("invalid KubernetesObjRef:\ngot  %s\nwant %s", kor, korWant)
	}
}

func TestKubernetesGetLabels(t *testing.T) {
	kor := KubernetesObjRef{
		ObjectReference: validKubernetesObjReference,
		BodyObj:         validKubernetesBodyObj,
	}
	labelsGot, err := kor.GetLabels()
	if err != nil {
		t.Fatal("invalid request:", err)
	}
	labelsWant := map[string]string{
		"app": "redis-controller",
		"com.docker.compose.service": "redis-controller",
	}
	if !reflect.DeepEqual(labelsGot, labelsWant) {
		t.Errorf("expected labels are not equal:\ngot  %s\nwant %s", labelsGot, labelsWant)
	}
}

func TestKubernetesMarshal2JSONStr(t *testing.T) {
	var powerStripReq PowerstripRequest
	err := DecodeRequest([]byte(validKubernetesSimpleRequest), &powerStripReq)
	if err != nil {
		t.Fatal("invalid request message:", err)
	}
	var kor KubernetesObjRef
	if err := powerStripReq.UnmarshalKubernetesObjRefClientBody(&kor); err != nil {
		t.Fatal("invalid validKubernetesRequestWoutEscQuot:", err)
	}
	_, err = kor.Marshal2JSONStr()
	if err != nil {
		t.Fatal("invalid KubernetesObjRef:", err)
	}
}

func TestMergeWithOverwriteKubernetes(t *testing.T) {
	pr := PowerstripRequest{
		ClientRequest: ClientRequest{
			Body: validKubernetesRequestWoutEscQuot,
		},
	}
	var kor KubernetesObjRef
	if err := pr.UnmarshalKubernetesObjRefClientBody(&kor); err != nil {
		t.Fatal("invalid validKubernetesRequestWoutEscQuot:", err)
	}
	korWant := KubernetesObjRef{
		ObjectReference: validKubernetesObjReference,
		BodyObj:         validKubernetesBodyObj,
	}
	kor.MergeWithOverwrite(korWant)
	if !reflect.DeepEqual(kor, korWant) {
		t.Errorf("invalid KubernetesObjRef:\ngot  %s\nwant %s", kor, korWant)
	}
}
