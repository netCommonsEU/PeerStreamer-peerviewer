import EError from '../utils/error';

export class HTTPErrorCode extends EError {
  constructor(code, expected) {
    super(`Expected code ${expected}, received ${code}`);
    this.code = code;
  }
}

export class ConnectionError extends EError {
  constructor(message = 'An error happened') {
    super(message);
  }
}
