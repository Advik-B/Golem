// Trailing comma in compound
try {
    nbt.parse('{key:"value",}'); // trailing comma
    test.fail("Parsing invalid SNBT should have thrown an error (trailing comma in compound).");
} catch (e) {
    test.log("Successfully caught expected error (trailing comma in compound): " + e);
}
// Invalid type (unquoted string)
try {
    nbt.parse('{key:unquotedString}');
    test.fail("Parsing invalid SNBT should have thrown an error (unquoted string).");
} catch (e) {
    test.log("Successfully caught expected error (unquoted string): " + e);
}
// Deeply nested invalid structure (trailing comma in nested list)
try {
    nbt.parse('{outer: {inner: [1, 2,]}}');
    test.fail("Parsing invalid SNBT should have thrown an error (trailing comma in nested list).");
} catch (e) {
    test.log("Successfully caught expected error (trailing comma in nested list): " + e);
}
// Incorrect array syntax (trailing comma in array)
try {
    nbt.parse('[I; 1, 2,]');
    test.fail("Parsing invalid SNBT should have thrown an error (trailing comma in array).");
} catch (e) {
    test.log("Successfully caught expected error (trailing comma in array): " + e);
}
// Unclosed bracket
try {
    nbt.parse('{key: 1');
    test.fail("Parsing invalid SNBT should have thrown an error (unclosed bracket).");
} catch (e) {
    test.log("Successfully caught expected error (unclosed bracket): " + e);
}