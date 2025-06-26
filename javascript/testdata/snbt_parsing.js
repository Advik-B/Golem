const snbt = '{ "key": "value", list: [1b, 2b, 3b] }';
const tag = nbt.parse(snbt);
if (tag.get('key').value !== 'value') test.fail("Parsed string value is incorrect");
if (tag.get('list').length !== 3) test.fail("Parsed list length is incorrect");

const pretty = tag.toPretty();
// FIX: The printer outputs unquoted keys if they are simple. The check must be more robust.
// This regex checks for 'key', optional whitespace, ':', optional whitespace, and then '"value"'.
if (!/key\s*:\s*"value"/.test(pretty)) {
    test.fail("Pretty print is missing key-value pair. Output was:\n" + pretty);
}
test.log("Pretty SNBT:\n" + pretty);
