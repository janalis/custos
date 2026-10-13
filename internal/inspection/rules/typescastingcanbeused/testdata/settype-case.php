<?php
function normalise($qty, $tags, $flag, $ratio) {
    <weak_warning descr="Use '$qty = (int)$qty' instead (a cast is clearer and faster).">settype($qty, 'INT')</weak_warning>;
    <weak_warning descr="Use '$tags = (array)$tags' instead (a cast is clearer and faster).">settype($tags, "Array")</weak_warning>;
    <weak_warning descr="Use '$flag = (bool)$flag' instead (a cast is clearer and faster).">SetType($flag, 'Boolean')</weak_warning>;
    settype($ratio, 'NULL');
    settype($ratio, 'Object');
}
