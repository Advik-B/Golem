const compound = nbt.newCompound();
compound.set("long", nbt.newLong(1234567890));
compound.set("string", nbt.newString("hello world"));

const binary = compound.write("rootName");
const result = nbt.read(binary);

if (result.name !== "rootName") test.fail("Root name was lost in translation. Got: " + result.name);
if (!nbt.compare(compound, result.tag, false)) test.fail("Tag is not the same after round trip");