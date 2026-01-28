package sdk

import "fmt"

func ExampleNewClient() {
	client := NewClient("https://api.example.com", "api-key")
	fmt.Println(client != nil)
	// Output: true
}

func ExampleVolumeAttachmentInput() {
	attachment := VolumeAttachmentInput{
		VolumeID:  "vol-123",
		MountPath: "/data",
	}
	fmt.Printf("%s:%s\n", attachment.VolumeID, attachment.MountPath)
	// Output: vol-123:/data
}

func ExampleLifecycleRule() {
	rule := LifecycleRule{
		BucketName:     "logs",
		Prefix:         "app/",
		ExpirationDays: 30,
		Enabled:        true,
	}
	fmt.Printf("%s:%s:%d:%t\n", rule.BucketName, rule.Prefix, rule.ExpirationDays, rule.Enabled)
	// Output: logs:app/:30:true
}

func ExampleCreateClusterInput() {
	input := CreateClusterInput{
		Name:        "dev-cluster",
		Version:     "1.29",
		WorkerCount: 3,
		HA:          true,
	}
	fmt.Printf("%s:%s:%d:%t\n", input.Name, input.Version, input.WorkerCount, input.HA)
	// Output: dev-cluster:1.29:3:true
}

func ExampleScaleClusterInput() {
	input := ScaleClusterInput{Workers: 5}
	fmt.Printf("%d\n", input.Workers)
	// Output: 5
}
