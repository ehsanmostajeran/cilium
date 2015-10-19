package kubernetes_policy

import (
	"encoding/json"
	"reflect"
	"testing"
)

var (
	bodyObjJson = `{"kind":"Pod","apiVersion":"v1","metadata":` +
		`{"name":"redis-controller-tester","namespace":"default",` +
		`"labels":{"app":"redis-controller","com.docker.compose.service":` +
		`"redis-controller"}},"spec":{"containers":[{"name":"redis","image":` +
		`"redis","ports":[{"containerPort":6379,"protocol":"TCP"}],"resources":` +
		`{},"terminationMessagePath":"/dev/termination-log","imagePullPolicy":` +
		`"IfNotPresent"}],"restartPolicy":"Always","dnsPolicy":"ClusterFirst"},` +
		`"status":{}}`
	orJson = `{"kind":"Pod","apiVersion":"v1"}`
)

func TestBodyObjOverwriteWith(t *testing.T) {
	var bo1, bo2, bowant BodyObj
	if err := json.Unmarshal([]byte(bodyObjJson), &bowant); err != nil {
		t.Fatalf("error while unmarshalling bodyObjJson: %s", err)
	}
	if err := json.Unmarshal([]byte(bodyObjJson), &bo1); err != nil {
		t.Fatalf("error while unmarshalling bodyObjJson: %s", err)
	}
	if err := json.Unmarshal([]byte(bodyObjJson), &bo2); err != nil {
		t.Fatalf("error while unmarshalling bodyObjJson: %s", err)
	}
	bo1["namespace"] = "foo"
	bo2["name"] = "bar"
	bo2["resourceVersion"] = "resourcebar"
	bo2["namespace"] = "foobar"
	bowant["namespace"] = "foobar"
	bowant["name"] = "bar"
	bowant["resourceVersion"] = "resourcebar"
	if err := bo1.OverwriteWith(bo2); err != nil {
		t.Errorf("error while executing OverwriteWith of BodyObj: %s", err)
	}
	if !reflect.DeepEqual(bo1, bowant) {
		t.Errorf("invalid BodyObj gotten:\ngot  %s\nwant %s\n", bo1, bowant)
	}
}

func TestObjectReferenceOverwriteWith(t *testing.T) {
	var or1, or2, orwant ObjectReference
	if err := json.Unmarshal([]byte(orJson), &orwant); err != nil {
		t.Fatalf("error while unmarshalling orJson: %s", err)
	}
	if err := json.Unmarshal([]byte(orJson), &or1); err != nil {
		t.Fatalf("error while unmarshalling orJson: %s", err)
	}
	if err := json.Unmarshal([]byte(orJson), &or2); err != nil {
		t.Fatalf("error while unmarshalling orJson: %s", err)
	}
	or1.Namespace = "foo"
	or2.Name = "bar"
	or2.ResourceVersion = "resourcebar"
	or2.Namespace = ""
	orwant.Namespace = "foo"
	orwant.Name = "bar"
	orwant.ResourceVersion = "resourcebar"
	if err := or1.OverwriteWith(or2); err != nil {
		t.Errorf("error while executing OverwriteWith of ObjectReference: %s", err)
	}
	if !reflect.DeepEqual(or1, orwant) {
		t.Errorf("invalid ObjectReference gotten:\ngot  %s\nwant %s\n", or1, orwant)
	}
}

func TestKubernetesConfigMergeWithOverwrite(t *testing.T) {
	var or ObjectReference
	if err := json.Unmarshal([]byte(orJson), &or); err != nil {
		t.Fatalf("error while unmarshalling orJson: %s", err)
	}
	var bo BodyObj
	if err := json.Unmarshal([]byte(bodyObjJson), &bo); err != nil {
		t.Fatalf("error while unmarshalling bodyObjJson: %s", err)
	}
	kc1 := KubernetesConfig{
		Priority:        222,
		ObjectReference: or,
		BodyObj:         bo,
	}
	kc2 := KubernetesConfig{
		Priority:        2,
		ObjectReference: or,
		BodyObj:         bo,
	}
	kc2.ObjectReference.Namespace = "foo"
	kcwant := KubernetesConfig{
		Priority:        2,
		ObjectReference: or,
		BodyObj:         bo,
	}
	kcwant.ObjectReference.Namespace = "foo"
	if err := kc1.MergeWithOverwrite(kc2); err != nil {
		t.Errorf("error while MergeWithOverwrite kc1 with kc2: %s", err)
	}
	if !reflect.DeepEqual(kc1, kcwant) {
		t.Errorf("invalid kc1:\ngot  %+v\nwant %+v", kc1, kcwant)
	}
}

func TestKubernetesConfigOverwriteWith(t *testing.T) {
	var or ObjectReference
	if err := json.Unmarshal([]byte(orJson), &or); err != nil {
		t.Fatalf("error while unmarshalling orJson: %s", err)
	}
	var bo BodyObj
	if err := json.Unmarshal([]byte(bodyObjJson), &bo); err != nil {
		t.Fatalf("error while unmarshalling bodyObjJson: %s", err)
	}
	kc1 := KubernetesConfig{
		Priority:        222,
		ObjectReference: or,
		BodyObj:         bo,
	}
	kc2 := KubernetesConfig{
		Priority:        2,
		ObjectReference: or,
		BodyObj:         bo,
	}
	kc1.ObjectReference.Name = "specialName"
	kc2.ObjectReference.Namespace = "specialNamespace"
	kc2.ObjectReference.ResourceVersion = ""
	kcwant := KubernetesConfig{
		Priority:        2,
		ObjectReference: or,
		BodyObj:         bo,
	}
	kcwant.ObjectReference.Name = "specialName"
	kcwant.ObjectReference.Namespace = "specialNamespace"
	kcwant.ObjectReference.ResourceVersion = ""
	if err := kc1.OverwriteWith(kc2); err != nil {
		t.Errorf("error while OverwriteWith kc1 with kc2: %s", err)
	}
	if !reflect.DeepEqual(kc1, kcwant) {
		t.Errorf("invalid kc1:\ngot  %+v\nwant %+v", kc1, kcwant)
	}
}

func TestOrderKubernetesConfigsByAscendingPriority(t *testing.T) {
	var or ObjectReference
	if err := json.Unmarshal([]byte(orJson), &or); err != nil {
		t.Fatalf("error while unmarshalling orJson: %s", err)
	}
	var bo BodyObj
	if err := json.Unmarshal([]byte(bodyObjJson), &bo); err != nil {
		t.Fatalf("error while unmarshalling bodyObjJson: %s", err)
	}
	kcs := []KubernetesConfig{
		KubernetesConfig{
			Priority:        222,
			ObjectReference: or,
			BodyObj:         bo,
		},
		KubernetesConfig{
			Priority:        2,
			ObjectReference: or,
			BodyObj:         bo,
		},
		KubernetesConfig{
			Priority:        1,
			ObjectReference: or,
			BodyObj:         bo,
		},
		KubernetesConfig{
			Priority:        199,
			ObjectReference: or,
			BodyObj:         bo,
		},
	}
	want := []KubernetesConfig{
		KubernetesConfig{
			Priority:        1,
			ObjectReference: or,
			BodyObj:         bo,
		},
		KubernetesConfig{
			Priority:        2,
			ObjectReference: or,
			BodyObj:         bo,
		},
		KubernetesConfig{
			Priority:        199,
			ObjectReference: or,
			BodyObj:         bo,
		},
		KubernetesConfig{
			Priority:        222,
			ObjectReference: or,
			BodyObj:         bo,
		},
	}
	OrderKubernetesConfigsByAscendingPriority(kcs)
	for i := range kcs {
		if kcs[i].Priority != want[i].Priority {
			t.Errorf("KubernetesConfigs are blady sorted (Priority):\ngot  %d\nwant %d", kcs[i].Priority, want[i].Priority)
		}
	}
}
