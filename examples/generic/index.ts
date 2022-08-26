// Code generated by tygo. DO NOT EDIT.

//////////
// source: generic.go

/**
 * Comment for UnionType
 */
export type UnionType = 
    /**
     * Comment for fields are possible
     */
    number /* uint64 */ | string | boolean | undefined // comment after
;
export type Derived = 
    number /* int */ | string // Line comment
;
export type Any = 
    string | any;
export type Empty = any;
export type Something = any;
export interface EmptyStruct {
}
export interface ValAndPtr<V extends any, PT extends (V | undefined), Unused extends number /* uint64 */> {
  Val: V;
  /**
   * Comment for ptr field
   */
  Ptr: PT; // ptr line comment
}
export interface ABCD<A extends string, B extends string, C extends UnionType, D extends number /* int64 */ | boolean> {
  a: A;
  b: B;
  c: C;
  d: D;
}
export interface Foo<A extends string | number /* uint64 */, B extends (A | undefined)> {
  Bar: A;
  Boo: B;
}
export interface WithFooGenericTypeArg<A extends Foo<string, string | undefined>> {
  some_field: A;
}
export interface Single<S extends string | number /* uint */> {
  Field: S;
}
export type SingleSpecific = Single<string>;