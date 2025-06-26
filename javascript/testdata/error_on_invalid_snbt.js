try {
    nbt.parse('{key:"value",}'); // trailing comma
    test.fail("Parsing invalid SNBT should have thrown an error.");
} catch (e) {
    test.log("Successfully caught expected error: " + e);
}