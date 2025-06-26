/**
 * NBT module bindings for the Golem server's Goja runtime.
 * This module allows JavaScript plugins to create, read, write,
 * and manipulate Minecraft NBT data structures.
 */
declare module "nbt" {
    /**
     * Base interface for all NBT tags.
     * Returned from factory functions and parsing functions.
     */
    export type Tag = {
        /**
         * Returns the numeric ID of the tag (e.g., 1 = Byte, 10 = Compound).
         */
        id(): number;

        /**
         * Returns the human-readable name of the tag type (e.g., "TAG_Byte").
         */
        typeName(): string;

        /**
         * Serializes the tag to compact SNBT (e.g., `{foo:123b}`).
         */
        toString(): string;

        /**
         * Serializes the tag to formatted, pretty-printed SNBT for debugging or logging.
         */
        toPretty(): string;

        /**
         * Returns a deep copy of the tag.
         */
        copy(): Tag;

        /**
         * Serializes the tag to a binary `ArrayBuffer` (uncompressed NBT format).
         * @param name Optional name of the root tag.
         */
        write(name?: string): ArrayBuffer;

        /**
         * Serializes the tag to a GZIP-compressed binary `ArrayBuffer`.
         * @param name Optional name of the root tag.
         */
        writeCompressed(name?: string): ArrayBuffer;
    };

    // ---------------------------------------
    // Core Functions
    // ---------------------------------------

    /**
     * Parses a stringified SNBT representation into a Tag.
     * The top-level tag is typically a CompoundTag.
     * @param snbt The SNBT string (e.g., `{foo: 123b, name: "Steve"}`).
     * @throws An error if parsing fails.
     */
    export function parse(snbt: string): Tag;

    /**
     * Converts a tag to a pretty-printed SNBT string.
     * @param tag The NBT tag to stringify.
     */
    export function toPretty(tag: Tag): string;

    /**
     * Converts a tag to compact SNBT (no line breaks or indentation).
     * @param tag The NBT tag to stringify.
     */
    export function toCompact(tag: Tag): string;

    /**
     * Reads binary NBT data from an ArrayBuffer (uncompressed).
     * @param buffer The buffer to read from (usually from `write()`).
     * @returns An object with the root name and parsed tag.
     */
    export function read(buffer: ArrayBuffer): { name: string; tag: Tag };

    /**
     * Reads compressed (GZIP) NBT data from an ArrayBuffer.
     * @param buffer The compressed buffer (from `writeCompressed()`).
     * @returns An object with the root name and parsed tag.
     */
    export function readCompressed(buffer: ArrayBuffer): { name: string; tag: Tag };

    /**
     * Compares two NBT tags for equality.
     * @param a The first tag.
     * @param b The second tag.
     * @param partial If true, checks that `a` is a subset of `b`.
     */
    export function compare(a: Tag, b: Tag, partial: boolean): boolean;

    // ---------------------------------------
    // Factory Functions
    // ---------------------------------------

    /**
     * Creates a new TAG_Byte.
     * @param v A signed 8-bit integer.
     */
    export function newByte(v: number): Tag;

    /**
     * Creates a new TAG_Short.
     * @param v A signed 16-bit integer.
     */
    export function newShort(v: number): Tag;

    /**
     * Creates a new TAG_Int.
     * @param v A signed 32-bit integer.
     */
    export function newInt(v: number): Tag;

    /**
     * Creates a new TAG_Long.
     * @param v A signed 64-bit integer.
     */
    export function newLong(v: number): Tag;

    /**
     * Creates a new TAG_Float.
     * @param v A 32-bit floating point number.
     */
    export function newFloat(v: number): Tag;

    /**
     * Creates a new TAG_Double.
     * @param v A 64-bit floating point number.
     */
    export function newDouble(v: number): Tag;

    /**
     * Creates a new TAG_String.
     * @param v A UTF-8 encoded string.
     */
    export function newString(v: string): Tag;

    /**
     * Creates a new TAG_Byte_Array.
     * @param v An Int8Array of signed 8-bit integers.
     */
    export function newByteArray(v: Int8Array): Tag;

    /**
     * Creates a new TAG_Int_Array.
     * @param v An Int32Array of signed 32-bit integers.
     */
    export function newIntArray(v: Int32Array): Tag;

    /**
     * Creates a new TAG_Long_Array.
     * @param v A BigInt64Array of signed 64-bit integers.
     */
    export function newLongArray(v: BigInt64Array): Tag;

    /**
     * Creates a new TAG_List. Items can be added via `.add()`.
     */
    export function newList(): Tag;

    /**
     * Creates a new TAG_Compound. Values can be added via `.set(key, value)`.
     */
    export function newCompound(): Tag;
}
