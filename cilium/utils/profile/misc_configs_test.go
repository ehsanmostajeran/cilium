package profile

import (
	"testing"
)

func TestCovering(t *testing.T) {
	test := func(coveringLabels map[string]string, labelsFromUser map[string]string, wanted bool) {
		coverage := Coverage{Labels: coveringLabels}
		if coverage.Covers(labelsFromUser) != wanted {
			t.Logf("converingLabels: %+v\n", coveringLabels)
			t.Logf("labelsFromUser : %+v\n", labelsFromUser)
			t.Errorf("Expected %+v, got %+v\n", wanted, !wanted)
		}
	}
	test(
		map[string]string{`com.intent.service`: "svc_dns"},
		map[string]string{`com.intent.service`: "svc_dns"},
		true,
	)
	test(
		map[string]string{`com\.intent\.service`: "svc_dns"},
		map[string]string{`comaintentaservice`: "svc_dns"},
		false,
	)
	test(
		map[string]string{`com.intent.service`: "svc_dns"},
		map[string]string{`comaintentaservice`: "svc_dns"},
		false,
	)
	test(
		map[string]string{
			`com\.intent\.service`: "svc_dns",
			`\.*production`:        "svc_dns",
		},
		map[string]string{
			`com.production`: "svc_dns",
		},
		false,
	)
	test(
		map[string]string{
			`com\.intent\.service`: "svc_dns",
			`\.*production`:        "^dns$",
		},
		map[string]string{
			`com.production`: "svc_dns",
		},
		false,
	)
	test(
		map[string]string{
			`com.intent.service`: `\.*svc_dns`,
			`\.*production`:      "dns",
		},
		map[string]string{
			`com.intent.service`: "foo_svc_dns_bar",
		},
		true,
	)
	test(
		map[string]string{
			`.*`:            `.*`,
			`\.*production`: "dns",
		},
		map[string]string{
			`com.intent.service`: "foo_svc_dns_bar",
		},
		false,
	)
}
