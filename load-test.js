import http from "k6/http";
import { check, sleep } from "k6";

export let options = {
  stages: [
    { duration: "30s", target: 10 },
    { duration: "1m", target: 10 },
    { duration: "30s", target: 0 },
  ],
  thresholds: {
    http_req_duration: ["p(99.9)<300"],
    http_req_failed: ["rate<0.001"],
  },
};

const BASE_URL = "http://localhost:8080";

function uuidv4() {
  return "xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx".replace(/[xy]/g, function (c) {
    var r = (Math.random() * 16) | 0,
      v = c == "x" ? r : (r & 0x3) | 0x8;
    return v.toString(16);
  });
}

// Основная функция виртуального пользователя
export default function () {
  // Генерируем уникальные ID для этой итерации
  const teamId = uuidv4();
  const teamName = `team_${teamId}`;
  const userId1 = `user_${uuidv4()}_1`;
  const userId2 = `user_${uuidv4()}_2`;
  const userId3 = `user_${uuidv4()}_3`;
  const prId = `pr_${uuidv4()}`;

  let res = http.get(BASE_URL + "/health");
  check(res, { "status is 200": (r) => r.status === 200 });

  res = http.post(
    BASE_URL + "/team/add",
    JSON.stringify({
      team_name: teamName,
      members: [
        {
          user_id: userId1,
          username: `alice_${teamId}`,
          is_active: true,
        },
        {
          user_id: userId2,
          username: `bob_${teamId}`,
          is_active: true,
        },
        {
          user_id: userId3,
          username: `charlie_${teamId}`,
          is_active: true,
        },
      ],
    }),
    {
      headers: { "Content-Type": "application/json" },
    },
  );
  check(res, { "team created": (r) => r.status === 200 || r.status === 201 });

  res = http.get(BASE_URL + `/team/get?team_name=${teamName}`);
  check(res, { "team fetched": (r) => r.status === 200 });

  res = http.post(
    BASE_URL + "/users/setIsActive",
    JSON.stringify({
      user_id: userId1,
      is_active: false,
    }),
    {
      headers: { "Content-Type": "application/json" },
    },
  );
  check(res, { "user status updated": (r) => r.status === 200 });

  res = http.post(
    BASE_URL + "/pullRequest/create",
    JSON.stringify({
      pull_request_id: prId,
      pull_request_name: `Feature: Add new API endpoint ${teamId}`,
      author_id: userId2,
    }),
    {
      headers: { "Content-Type": "application/json" },
    },
  );
  check(res, { "PR created": (r) => r.status === 200 || r.status === 201 });

  res = http.get(BASE_URL + `/users/getReview?user_id=${userId3}`);
  check(res, { "review fetched": (r) => r.status === 200 });

  res = http.post(
    BASE_URL + "/pullRequest/merge",
    JSON.stringify({
      pull_request_id: prId,
    }),
    {
      headers: { "Content-Type": "application/json" },
    },
  );
  check(res, { "PR merged": (r) => r.status === 200 });

  sleep(1);
}
