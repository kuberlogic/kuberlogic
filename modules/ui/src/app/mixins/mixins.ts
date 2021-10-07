/*
 * CloudLinux Software Inc 2019-2021 All Rights Reserved
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

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
