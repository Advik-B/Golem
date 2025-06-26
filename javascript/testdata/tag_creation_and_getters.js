const comp = nbt.newCompound();
comp.set("b", nbt.newByte(127));
if (comp.get("b").value !== 127) test.fail("Byte value mismatch.");
if (comp.get("b").typeName() !== "TAG_Byte") test.fail("Byte type name mismatch.");
