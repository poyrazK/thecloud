# ADR 021: Storage Integrity and Orphan Cleanup

## Status
Accepted

## Context
Following the transition to streaming storage (ADR 020), two operational challenges were identified:
1.  **Orphaned Data**: Failed or interrupted uploads could leave physical files on storage nodes without corresponding metadata in the database, or vice versa. Since streaming writes start immediately, a failure halfway through would leave a partial file that was difficult to identify and clean up.
2.  **Incomplete Metadata**: Content-Type detection relied on client headers, which are often missing or incorrect. Additionally, there was no mechanism to verify the end-to-end integrity of a file after it had been successfully streamed and assembled.

## Decision
We implemented a robust two-phase upload process combined with automated metadata enrichment and data integrity verification.

### 1. Two-Phase Upload Pattern (Atomic Metadata)
-   Objects now follow a state machine: `PENDING` -> `AVAILABLE`.
-   **Step 1**: Create a metadata record with `upload_status = 'PENDING'` before any physical data is written.
-   **Step 2**: Stream data to the storage backend.
-   **Step 3**: On success, update the record to `AVAILABLE` and finalize size/checksum.
-   This ensures that every physical write attempt has a corresponding database entry, making "invisible" orphans impossible.

### 2. Automated Orphan Garbage Collection
-   Updated the `StorageCleanupWorker` to handle failed uploads.
-   The worker identifies `PENDING` records older than a configured threshold (default 1 hour).
-   It deletes the corresponding physical files from storage nodes and then removes the metadata from the database.
-   This provides a self-healing mechanism for failed uploads and partial writes.

### 3. Streaming Content-Type Sniffing
-   Implemented automatic MIME type detection in the `StorageService`.
-   The service buffers only the first 512 bytes of the stream to use `http.DetectContentType`.
-   The rest of the stream is then transparently combined with the buffer using `io.MultiReader`, maintaining constant memory usage.

### 4. End-to-End Data Integrity (SHA-256 Checksums)
-   The `StorageService` now calculates a SHA-256 hash of the data stream in real-time using `io.TeeReader`.
-   The final checksum is stored in the `objects` table upon successful upload completion.
-   This provides a "gold standard" for data integrity that can be used for periodic audits or by clients to verify downloads.

## Consequences
-   **Reliability**: Interrupted uploads are now automatically cleaned up, preventing storage leakage.
-   **Observability**: Pending uploads are visible in the database, allowing for better monitoring of active system state.
-   **Data Quality**: Objects now have accurate Content-Types and verifiable checksums regardless of client-provided headers.
-   **Performance**: Minimal overhead added; content-sniffing only buffers 512 bytes, and checksumming is performed in a single pass during the write operation.
-   **Database**: Added `upload_status` and `checksum` columns to the `objects` table, with an index on `pending` objects to optimize worker performance.
