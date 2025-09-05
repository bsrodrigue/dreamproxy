document.addEventListener("DOMContentLoaded", async () => {
  // const MEDIA_SOURCE = new MediaSource();
  // const VIDEO_PLAYER = document.getElementById("videoPlayer");
  // const FILE_NAME = "episode01.mp4";
  // const CHUNK_SIZE = 1 * 1024 * 1024; // Reduced to 1MB chunks
  //
  // // Set up video player
  // VIDEO_PLAYER.src = URL.createObjectURL(MEDIA_SOURCE);
  //
  // VIDEO_PLAYER.addEventListener("click", () => {
  //   VIDEO_PLAYER.muted = true;
  //   VIDEO_PLAYER.play().catch((err) => console.error("Playback failed:", err));
  // });
  //
  // // First, let's detect the actual codec of the video file
  // async function detectVideoCodec() {
  //   try {
  //     // Try to get a small chunk to analyze
  //     const response = await fetch(`/${FILE_NAME}`, {
  //       headers: { "Range": "bytes=0-1023" },
  //       method: "GET"
  //     });
  //
  //     if (!response.ok) {
  //       throw new Error(`Failed to fetch video header: ${response.status}`);
  //     }
  //
  //     // Common codec combinations to try
  //     const codecsToTry = [
  //       'video/mp4; codecs="avc1.42E01E, mp4a.40.2"', // H.264 Baseline + AAC
  //       'video/mp4; codecs="avc1.4D401E, mp4a.40.2"', // H.264 Main + AAC
  //       'video/mp4; codecs="avc1.64001E, mp4a.40.2"', // H.264 High + AAC
  //       'video/mp4; codecs="mp4v.20.9, mp4a.40.2"',   // MPEG-4 + AAC
  //       'video/mp4; codecs="avc1.42001E, mp4a.40.5"', // H.264 + HE-AAC
  //       'video/mp4',                                   // Generic MP4
  //     ];
  //
  //     for (const codec of codecsToTry) {
  //       if (MediaSource.isTypeSupported(codec)) {
  //         console.log(`Supported codec found: ${codec}`);
  //         return codec;
  //       }
  //     }
  //
  //     throw new Error('No supported codec found for this video file');
  //
  //   } catch (error) {
  //     console.error('Error detecting codec:', error);
  //     // Fallback to most common codec
  //     return 'video/mp4; codecs="avc1.42E01E, mp4a.40.2"';
  //   }
  // }
  //
  // async function getFileSize() {
  //   try {
  //     const response = await fetch(`/${FILE_NAME}`, {
  //       method: "HEAD"
  //     });
  //
  //     if (!response.ok) {
  //       throw new Error(`Failed to get file size: ${response.status}`);
  //     }
  //
  //     const contentLength = response.headers.get("Content-Length");
  //     if (!contentLength) {
  //       throw new Error('Content-Length header missing - server may not support range requests');
  //     }
  //
  //     return parseInt(contentLength);
  //   } catch (error) {
  //     console.error('Error getting file size:', error);
  //     throw error;
  //   }
  // }
  //
  // async function loadChunkBuffer(chunkIndex, chunkCount, fileSize) {
  //   const lowerBound = chunkIndex * CHUNK_SIZE;
  //   const upperBound = Math.min(lowerBound + CHUNK_SIZE - 1, fileSize - 1);
  //   const range = `bytes=${lowerBound}-${upperBound}`;
  //
  //   console.log(`Fetching chunk ${chunkIndex + 1}/${chunkCount}: ${range}`);
  //
  //   try {
  //     const response = await fetch(`/${FILE_NAME}`, {
  //       headers: { "Range": range },
  //       method: "GET"
  //     });
  //
  //     if (!response.ok) {
  //       throw new Error(`Failed to fetch chunk ${chunkIndex}: ${response.status} ${response.statusText}`);
  //     }
  //
  //     // Verify we got a partial content response
  //     if (response.status !== 206) {
  //       console.warn(`Expected 206 Partial Content, got ${response.status} - server may not support range requests properly`);
  //     }
  //
  //     const buffer = await response.arrayBuffer();
  //     console.log(`Chunk ${chunkIndex + 1} loaded: ${buffer.byteLength} bytes`);
  //     return buffer;
  //
  //   } catch (error) {
  //     console.error(`Error loading chunk ${chunkIndex}:`, error);
  //     throw error;
  //   }
  // }
  //
  // MEDIA_SOURCE.addEventListener("sourceopen", async () => {
  //   console.log("MediaSource opened");
  //
  //   try {
  //     // Detect and use appropriate codec
  //     const codec = await detectVideoCodec();
  //     console.log(`Using codec: ${codec}`);
  //
  //     if (!MediaSource.isTypeSupported(codec)) {
  //       throw new Error(`Codec not supported: ${codec}`);
  //     }
  //
  //     const sourceBuffer = MEDIA_SOURCE.addSourceBuffer(codec);
  //     const fileSize = await getFileSize();
  //     const chunkCount = Math.ceil(fileSize / CHUNK_SIZE);
  //
  //     console.log(`File size: ${fileSize} bytes, ${chunkCount} chunks of ${CHUNK_SIZE} bytes each`);
  //
  //     let currentChunk = 0;
  //     let isLoading = false;
  //     let hasError = false;
  //
  //     async function loadNextChunk() {
  //       if (hasError || currentChunk >= chunkCount || isLoading) {
  //         return;
  //       }
  //
  //       if (sourceBuffer.updating) {
  //         console.log("SourceBuffer is updating, waiting...");
  //         setTimeout(loadNextChunk, 100);
  //         return;
  //       }
  //
  //       isLoading = true;
  //
  //       try {
  //         const chunkBuffer = await loadChunkBuffer(currentChunk, chunkCount, fileSize);
  //
  //         // Double-check sourceBuffer state before appending
  //         if (MEDIA_SOURCE.readyState !== 'open') {
  //           console.error(`MediaSource readyState is '${MEDIA_SOURCE.readyState}', cannot append buffer`);
  //           return;
  //         }
  //
  //         if (sourceBuffer.updating) {
  //           console.log("SourceBuffer became busy, retrying...");
  //           setTimeout(loadNextChunk, 100);
  //           return;
  //         }
  //
  //         sourceBuffer.appendBuffer(chunkBuffer);
  //         currentChunk++;
  //
  //       } catch (error) {
  //         console.error(`Failed to load chunk ${currentChunk}:`, error);
  //         hasError = true;
  //       } finally {
  //         isLoading = false;
  //       }
  //     }
  //
  //     sourceBuffer.addEventListener("updateend", () => {
  //       console.log(`Chunk ${currentChunk}/${chunkCount} successfully appended`);
  //
  //       if (hasError) {
  //         console.error("Stopping due to previous error");
  //         return;
  //       }
  //
  //       if (currentChunk < chunkCount) {
  //         // Load next chunk with a small delay to prevent overwhelming
  //         setTimeout(loadNextChunk, 50);
  //       } else {
  //         // All chunks loaded, end the stream
  //         try {
  //           if (MEDIA_SOURCE.readyState === 'open') {
  //             console.log("All chunks loaded, ending MediaSource stream");
  //             MEDIA_SOURCE.endOfStream();
  //           } else {
  //             console.log(`Cannot end stream, MediaSource readyState is '${MEDIA_SOURCE.readyState}'`);
  //           }
  //         } catch (error) {
  //           console.error("Error ending stream:", error);
  //         }
  //       }
  //     });
  //
  //     sourceBuffer.addEventListener("error", (event) => {
  //       console.error("SourceBuffer error:", event);
  //       hasError = true;
  //
  //       // Try to get more details about the error
  //       if (sourceBuffer.buffered.length > 0) {
  //         console.log("Buffered ranges:", Array.from({ length: sourceBuffer.buffered.length }, (_, i) =>
  //           `${sourceBuffer.buffered.start(i)}-${sourceBuffer.buffered.end(i)}`
  //         ).join(', '));
  //       }
  //     });
  //
  //     sourceBuffer.addEventListener("abort", () => {
  //       console.log("SourceBuffer operation aborted");
  //       hasError = true;
  //     });
  //
  //     // Start loading the first chunk
  //     await loadNextChunk();
  //
  //   } catch (error) {
  //     console.error("Error in sourceopen handler:", error);
  //   }
  // });
  //
  // MEDIA_SOURCE.addEventListener("sourceended", () => {
  //   console.log("MediaSource ended successfully");
  // });
  //
  // MEDIA_SOURCE.addEventListener("sourceclose", () => {
  //   console.log("MediaSource closed");
  //   URL.revokeObjectURL(VIDEO_PLAYER.src);
  // });
  //
  // // Enhanced video player error handling
  // VIDEO_PLAYER.addEventListener("error", (event) => {
  //   console.error("Video player error:", {
  //     error: VIDEO_PLAYER.error,
  //     code: VIDEO_PLAYER.error?.code,
  //     message: VIDEO_PLAYER.error?.message,
  //     networkState: VIDEO_PLAYER.networkState,
  //     readyState: VIDEO_PLAYER.readyState
  //   });
  // });
  //
  // VIDEO_PLAYER.addEventListener("loadstart", () => {
  //   console.log("Video load started");
  // });
  //
  // VIDEO_PLAYER.addEventListener("loadedmetadata", () => {
  //   console.log("Video metadata loaded", {
  //     duration: VIDEO_PLAYER.duration,
  //     videoWidth: VIDEO_PLAYER.videoWidth,
  //     videoHeight: VIDEO_PLAYER.videoHeight
  //   });
  // });
  //
  // VIDEO_PLAYER.addEventListener("canplay", () => {
  //   console.log("Video can start playing");
  // });
  //
  // VIDEO_PLAYER.addEventListener("canplaythrough", () => {
  //   console.log("Video can play through without buffering");
  // });
});
