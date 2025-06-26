declare module "nbt" {
    export type Tag = {
        id(): number;
        typeName(): string;
        toString(): string;
        toPretty(): string;
        copy(): Tag;
        write(name?: string): ArrayBuffer;
        writeCompressed(name?: string): ArrayBuffer;
    };

    export function parse(snbt: string): Tag;
    export function toPretty(tag: Tag): string;
    export function toCompact(tag: Tag): string;
    export function read(buffer: ArrayBuffer): { name: string; tag: Tag };
    export function readCompressed(buffer: ArrayBuffer): { name: string; tag: Tag };
    export function compare(a: Tag, b: Tag, partial: boolean): boolean;

    export function newByte(v: number): Tag;
    export function newShort(v: number): Tag;
    export function newInt(v: number): Tag;
    export function newLong(v: number): Tag;
    export function newFloat(v: number): Tag;
    export function newDouble(v: number): Tag;
    export function newString(v: string): Tag;
    export function newByteArray(v: number[]): Tag;
    export function newIntArray(v: number[]): Tag;
    export function newLongArray(v: number[]): Tag;
    export function newList(): Tag;
    export function newCompound(): Tag;
}
