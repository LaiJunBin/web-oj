import http from 'k6/http';
import { sleep } from 'k6';
export const options = {
  vus: 10,
  duration: '5s',
  thresholds: {
    http_req_duration: [`p(95)<${__ENV.HTTP_REQ_DURATION}`], // 95% 的請求需小於 500ms
  },
};

export default function () {
  http.get(`http://${__ENV.PROJECT_CONTAINER_NAME}/`);
  sleep(1);
}