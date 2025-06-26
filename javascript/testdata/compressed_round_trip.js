const tag = nbt.newIntArray([10, 20, 30]);
const compressedBinary = tag.writeCompressed("compressedRoot");
const result = nbt.readCompressed(compressedBinary);

if (result.name !== "compressedRoot") test.fail("Root name was lost in compressed translation.");
if (!nbt.compare(tag, result.tag, false)) test.fail("Tag is not the same after compressed round trip");
