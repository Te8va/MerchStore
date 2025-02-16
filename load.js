import http from 'k6/http';
import { check, group } from 'k6';
import { randomString, randomIntBetween } from 'https://jslib.k6.io/k6-utils/1.2.0/index.js';

export let options = {
    stages: [
        { duration: '1m', target: 750 },
        { duration: '2m', target: 1500 },
        { duration: '1m', target: 0 }
    ],
    thresholds: {
        http_reqs: ['rate>=1000'],
        http_req_failed: ['rate<=0.0001']
    }
};

let merch =['t-shirt','cup', 'book', 'pen', 'powerbank', 'hoody', 'umbrella', 'socks', 'wallet', 'pink-hoody'];

function auth(username, password) {
    const authUrl = 'http://localhost:8080/api/auth';

    const payload = JSON.stringify({
        username: username,
        password: password
    });

    const params = {
        headers: {
        'Content-Type': 'application/json'
        }
    };

    let response = http.post(authUrl, payload, params);

    check(response, {
        'login status is 200': (r) => r.status === 200
    });

    return response.json()['token'];
}

function buy(token, item) {
    const buyUrl = `http://localhost:8080/api/buy/${item}`;

    const params = {
        headers: {
        'Authorization': `Bearer ${token}`
        }
    };

    let response = http.get(buyUrl, params);

    check(response, {
        'buy status is 200 or not enough coins': (r) => [200, 400].includes(r.status)
    });

    return;
}

function sendCoin(token, toUser, amount) {
    const authUrl = 'http://localhost:8080/api/sendCoin';

    const payload = JSON.stringify({
        toUser: toUser,
        amount: amount
    });

    const params = {
        headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${token}`
        }
    };

    let response = http.post(authUrl, payload, params);

    check(response, {
        'send status is 200 or not enough coins': (r) => [200, 400].includes(r.status)
    });

    return;
}

function info(token) {
    const authUrl = 'http://localhost:8080/api/info';

    const params = {
        headers: {
        'Authorization': `Bearer ${token}`
        }
    };

    let response = http.get(authUrl, params);

    check(response, {
        'info is 200': (r) => r.status === 200
    });

    return;
}

export default function () {
    let token = auth(randomString(7),randomString(4));
    switch (randomIntBetween(1, 4)) {
        case 1:
            group('info', function () {
                info(token);
            });
            break;
        case 2:
            group('info', function () {
                info(token);
            });
            break;
        case 3:
            group('buy', function () {
                buy(token, merch[randomIntBetween(0,9)]);
            });
            break;
        case 4:
            group('send', function () {
                let toUser = randomString(7);
                auth(toUser,randomString(4));
                sendCoin(token, toUser, randomIntBetween(1,10));
            });
            break;
    }
}