try {
    const list = nbt.newList();
    list.add(nbt.newInt(1));
    list.add(nbt.newString("I should not be here")); // Mismatched type
    test.fail("Adding a mismatched type to a list should have thrown an error.");
} catch (e) {
    test.log("Successfully caught expected error: " + e);
}
