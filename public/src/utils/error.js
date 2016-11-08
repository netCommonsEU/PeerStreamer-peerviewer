export default class ExtensibleError {
  constructor(message) {
    this.name = this.constructor.name;
    this.message = message;
    /* istanbul ignore if  */
    if (typeof Error.captureStackTrace === 'function') {
      Error.captureStackTrace(this, this.constructor);
    } else {
      this.stack = (new Error(message)).stack;
    }
  }
}

ExtensibleError.prototype = Error.prototype;
