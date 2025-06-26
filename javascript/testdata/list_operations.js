const list = nbt.newList();
list.add(nbt.newInt(1));
list.add(nbt.newInt(2));
if (list.length !== 2) test.fail("List length should be 2");
if (list.get(1).value !== 2) test.fail("List element value is incorrect");
if (list.listType !== 3) test.fail("List type ID should be TAG_Int (3)");
