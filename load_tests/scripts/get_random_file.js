import http from "k6/http";
import { check, sleep } from "k6";

export let options = {
  duration: "1m",
  vus: 1000,
};

const randomFilesIds = JSON.parse(open("../dirs.json"));

export default function () {
  let randomFileId =
    randomFilesIds[Math.floor(Math.random() * randomFilesIds.length)];
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
