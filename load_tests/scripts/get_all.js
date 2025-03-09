import http from "k6/http";
import { check, sleep } from "k6";

const config = JSON.parse(open("../static/config.json"));

export let options = {
  duration: "1m",
  vus: 100,
};

export default function () {
  let randomPage = Math.floor(Math.random() * 100);
  let res = http.get(config.host + "/files?page=" + randomPage);
  check(res, {
    "status was 200": (r) => r.status === 200,
    "response time < 200ms": (r) => r.timings.duration < 200,
  });
  sleep(1);
}
