import { Schema, arrayOf } from 'normalizr';

export const stream = new Schema('streams');
export const arrayOfStreams = arrayOf(stream);
