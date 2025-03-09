import http from "k6/http";
import { check, sleep } from "k6";

export let options = {
  stages: [
    { duration: "5s", target: 500 },
    { duration: "1m", target: 500 },
  ],
  thresholds: {
    http_req_duration: [
      {
        threshold: "p(95)<1000",
        abortOnFail: true,
        delayAbortEval: "10s",
      },
    ],
  },
};

const imageFile = "large_image.jpg";
const file = open("../static/" + imageFile, "b");
const numChunks = 10;
const chunkSize = Math.ceil(file.byteLength / numChunks); // Divide into 5 chunks

const requestsData = [];

export function setup() {
  const chunks = [];
  for (let chunkNumber = 1; chunkNumber <= numChunks; chunkNumber++) {
    const start = (chunkNumber - 1) * chunkSize;
    const end = Math.min(chunkNumber * chunkSize, chunkSize);
    const chunk = file.slice(start, end);
    chunks.push(chunk);
  }

  for (let i = 0; i < chunks.length; i++) {
    const fileChunk = http.file(chunks[i], imageFile);
    requestsData.push([
      "POST",
      "http://localhost:8008/files",
      {
        file: fileChunk,
        fileId: null,
        chunkNumber: i + 1,
        totalChunks: numChunks,
        filename: imageFile,
      },
    ]);
  }
}

export default function () {
  let responses = http.batch(requestsData);
  for (let res of responses) {
    check(res, {
      "status was 200": (r) => r.status === 200,
      "response time < 500ms": (r) => r.timings.duration < 1000,
    });
  }
  sleep(1);
}
