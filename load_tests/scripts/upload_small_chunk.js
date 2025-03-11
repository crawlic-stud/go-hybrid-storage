import http from "k6/http";
import { check, sleep } from "k6";

export let options = {
  stages: [
    { duration: "1m", target: 1000 },
    { duration: "1m", target: 1000 },
  ],
  thresholds: {
    http_req_duration: [
      {
        threshold: "p(95)<1000",
        abortOnFail: true,
        delayAbortEval: "10s",
      },
    ],
    http_req_failed: [
      {
        threshold: "rate<0.05",
        abortOnFail: true,
        delayAbortEval: "10s",
      },
    ],
  },
};

const imageFile = "small_image.jpg";
const file = http.file(open("../static/" + imageFile, "b"), imageFile);
const formData = {
  file: file,
  fileId: null,
  chunkNumber: 1,
  totalChunks: 1,
  filename: imageFile,
};

export default function () {
  let res = http.post("http://localhost:8008/files", formData);
  check(res, {
    "status was 200": (r) => r.status === 200,
    "response time < 1000ms": (r) => r.timings.duration < 1000,
  });
  sleep(1);
}
