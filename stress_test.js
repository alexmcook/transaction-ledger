import http from 'k6/http';
import { check, sleep } from 'k6';
import { SharedArray } from 'k6/data';

const account_ids = new SharedArray('account_ids', function () {
  const f = open('./data/account_ids.csv');
  return f.split('\n').filter(line => line.trim() !== '');
});

export const options = {
  vus: 200,
  duration: '20s',

  // stages: [
  //   { duration: '30s', target: 50 },
  //   { duration: '2m', target: 50 },
  //   { duration: '30s', target: 0 },
  // ],

  thresholds: {
    http_req_failed: ['rate<0.01'],
    http_req_duration: ['p(95)<200'],
  },
};

export default function () {
  const idx = Math.floor(Math.random() * account_ids.length);
  const amount = Math.floor(Math.random() * 10000) + 100;
  const payload = JSON.stringify({
    accountId: account_ids[idx],
    type: amount > 0 ? 1 : 2,
    amount: amount
  });

  const params = {
    headers: { 'Content-Type': 'application/json' },
  };

  const res = http.post('http://localhost:8080/transactions', payload, params);

  check(res, {
    'is status 201': (r) => r.status === 201,
  });
}
