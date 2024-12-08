import http from "k6/http";
import { check, sleep } from "k6";

export let options = {
  duration: "1m",
  vus: 1000,
};

export default function () {
  let randomPage = Math.floor(Math.random() * 100);
  let res = http.get("http://localhost:8008/files?page=" + randomPage);
  check(res, {
    "status was 200": (r) => r.status === 200,
    "response time < 200ms": (r) => r.timings.duration < 200,
  });
  sleep(1);
}
