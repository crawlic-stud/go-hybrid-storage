import http from "k6/http";
import { check, sleep } from "k6";

export let options = {
  duration: "5s",
  vus: 1000,
};

const imageFile = "small_image.jpg";

const file = http.file(open("../" + imageFile, "b"), imageFile);

export default function () {
  const formData = { file: file };

  let res = http.post("http://localhost:8008/files", formData);
  check(res, {
    "status was 200": (r) => r.status === 200,
    "response time < 200ms": (r) => r.timings.duration < 200,
  });
  sleep(1);
}
