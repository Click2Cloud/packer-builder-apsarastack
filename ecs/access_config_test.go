package ecs

import (
	"os"
	"testing"
)

func testApsaraStackAccessConfig() *ApsaraStackAccessConfig {
	return &ApsaraStackAccessConfig{
		ApsaraStackAccessKey: "ak",
		ApsaraStackSecretKey: "acs",
	}

}

func TestApsaraStackAccessConfigPrepareRegion(t *testing.T) {
	c := testApsaraStackAccessConfig()

	c.ApsaraStackRegion = ""
	if err := c.Prepare(nil); err == nil {
		t.Fatalf("should have err")
	}

	c.ApsaraStackRegion = "cn-beijing"
	if err := c.Prepare(nil); err != nil {
		t.Fatalf("shouldn't have err: %s", err)
	}

	os.Setenv("APSARASTACK_REGION", "cn-hangzhou")
	c.ApsaraStackRegion = ""
	if err := c.Prepare(nil); err != nil {
		t.Fatalf("shouldn't have err: %s", err)
	}

	c.ApsaraStackAccessKey = ""
	if err := c.Prepare(nil); err == nil {
		t.Fatalf("should have err")
	}

	c.ApsaraStackProfile = "default"
	if err := c.Prepare(nil); err != nil {
		t.Fatalf("shouldn't have err: %s", err)
	}

	c.ApsaraStackProfile = ""
	os.Setenv("APSARASTACK_PROFILE", "default")
	if err := c.Prepare(nil); err != nil {
		t.Fatalf("shouldn't have err: %s", err)
	}

	c.ApsaraStackSkipValidation = false
}
