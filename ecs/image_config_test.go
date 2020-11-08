package ecs

import (
	"testing"
)

func testApsaraStackImageConfig() *ApsaraStackImageConfig {
	return &ApsaraStackImageConfig{
		ApsaraStackImageName: "foo",
	}
}

func TestECSImageConfigPrepare_name(t *testing.T) {
	c := testApsaraStackImageConfig()
	if err := c.Prepare(nil); err != nil {
		t.Fatalf("shouldn't have err: %s", err)
	}

	c.ApsaraStackImageName = ""
	if err := c.Prepare(nil); err == nil {
		t.Fatal("should have error")
	}
}

func TestAMIConfigPrepare_regions(t *testing.T) {
	c := testApsaraStackImageConfig()
	c.ApsaraStackImageDestinationRegions = nil
	if err := c.Prepare(nil); err != nil {
		t.Fatalf("shouldn't have err: %s", err)
	}

	c.ApsaraStackImageDestinationRegions = []string{"cn-beijing", "cn-hangzhou", "eu-central-1"}
	if err := c.Prepare(nil); err != nil {
		t.Fatalf("bad: %s", err)
	}

	c.ApsaraStackImageDestinationRegions = nil
	c.ApsaraStackImageSkipRegionValidation = true
	if err := c.Prepare(nil); err != nil {
		t.Fatal("shouldn't have error")
	}
	c.ApsaraStackImageSkipRegionValidation = false
}

func TestECSImageConfigPrepare_imageTags(t *testing.T) {
	c := testApsaraStackImageConfig()
	c.ApsaraStackImageTags = map[string]string{
		"TagKey1": "TagValue1",
		"TagKey2": "TagValue2",
	}
	if err := c.Prepare(nil); len(err) != 0 {
		t.Fatalf("err: %s", err)
	}
	if len(c.ApsaraStackImageTags) != 2 || c.ApsaraStackImageTags["TagKey1"] != "TagValue1" ||
		c.ApsaraStackImageTags["TagKey2"] != "TagValue2" {
		t.Fatalf("invalid value, expected: %s, actual: %s", map[string]string{
			"TagKey1": "TagValue1",
			"TagKey2": "TagValue2",
		}, c.ApsaraStackImageTags)
	}
}
