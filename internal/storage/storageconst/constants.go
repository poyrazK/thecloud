// Package storageconst centralises configuration constants that were
// previously duplicated across internal/storage/coordinator and
// internal/storage/node. Concentrating them here avoids drift — a chunk
// size or object size limit changed on one side that the other side does
// not know about silently breaks streaming/quorum behaviour.
package storageconst

// MaxObjectSize is the hard ceiling on a single object stored anywhere in
// the cluster. It guards against memory exhaustion on the streaming path
// (coordinator) and against unbounded file writes on the storage node.
//
// Keep this in sync with any service-layer MaxPartSize used by multipart
// uploads.
const MaxObjectSize = 5 * 1024 * 1024 * 1024 // 5 GB

// ChunkSize is the streaming buffer used both when the coordinator
// broadcasts an object to replicas and when storage nodes ship retrieve
// responses back. Changing this changes wire chunking, so it must stay
// identical on both sides.
const ChunkSize = 1024 * 1024 // 1 MB
