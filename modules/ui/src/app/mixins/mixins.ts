/**
 * stub-class for basing object class to create mixin, like:
 * `class RequestObjectCard extends ExtraCardMixin(BaseObject) {`
 */
export class BaseObject {
    // eslint-disable-next-line  no-empty, no-empty-function, @typescript-eslint/no-empty-function
    constructor(..._args: any[]) {}
}

/**
 * Helper for mixins creation
 * type helper for creating mixins, like:
 * `export function ExtraCardMixin<TBase extends Constructor>(Base: TBase) {`
 */
export type Constructor<T = BaseObject> = new(...args: any[]) => T;
