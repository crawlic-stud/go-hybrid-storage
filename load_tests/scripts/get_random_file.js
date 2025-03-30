import http from "k6/http";
import { check, sleep } from "k6";

export let options = {
  setupTimeout: "5m",
  stages: [
    { duration: "1m", target: 500 },
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
    http_req_failed: [
      {
        threshold: "rate<0.05",
        abortOnFail: true,
        delayAbortEval: "10s",
      },
    ],
  },
};

const smallImage = "small_image.jpg";
const smallFormData = {
  file: http.file(open("../static/" + smallImage, "b"), smallImage),
  fileId: null,
  chunkNumber: 1,
  totalChunks: 1,
  filename: smallImage,
};

const largeImage = "large_image.jpg";
const largeFormData = {
  file: http.file(open("../static/" + largeImage, "b"), largeImage),
  fileId: null,
  chunkNumber: 1,
  totalChunks: 1,
  filename: largeImage,
};

function randChoice(items) {
  return items[Math.floor(Math.random() * items.length)];
}

export function setup() {
  const fileIds = [];
  var smallInserted = 0;
  var largeInserted = 0;
  // create many files firstly
  for (let i = 0; i <= 100; i++) {
    var randImage = randChoice([smallFormData, largeFormData]);
    if (randImage == smallFormData) {
      smallInserted++;
    } else {
      largeInserted++;
    }
    var res = http.post("http://localhost:8008/files", randImage);
    fileIds.push(res.json().fileId);
  }
  console.log(
    `[!] INSERTED ${smallInserted} small and ${largeInserted} large files [!]`
  );
  return { fileIds: fileIds };
}

export default function (data) {
  let randomFileId = randChoice(data.fileIds);
  let res1 = http.get("http://localhost:8008/files/" + randomFileId);
  let res2 = http.get(
    "http://localhost:8008/files/" + randomFileId + "/metadata"
  );
  for (let res of [res1, res2]) {
    check(res, {
      "status was 200": (r) => r.status === 200,
      "response time < 500ms": (r) => r.timings.duration < 500,
    });
  }
  sleep(1);
}
