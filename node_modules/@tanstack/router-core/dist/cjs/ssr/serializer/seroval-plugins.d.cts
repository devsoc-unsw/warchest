import { RawStream } from './RawStream.cjs';
import { Plugin } from 'seroval';
export declare const defaultSerovalPlugins: (Plugin<Error, any> | Plugin<RawStream, any> | Plugin<ReadableStream<any>, any>)[];
