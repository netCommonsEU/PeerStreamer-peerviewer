import { HTTPErrorCode, ConnectionError } from './errors';
import * as endpoints from './endpoints';

export const fetchStreams = () => new Promise((resolve, reject) => {
  const req = new XMLHttpRequest();
  req.open('GET', endpoints.STREAM_LIST);
  req.onload = () => {
    if (200 <= req.status < 300) {
      try {
        const data = JSON.parse(req.responseText);
        resolve(data);
      } catch (e) {
        reject(e);
      }
    } else {
      reject(new HTTPErrorCode(req.status, '2xx'));
    }
  };
  req.onerror = () => {
    reject(new ConnectionError());
  };
  req.send();
});
