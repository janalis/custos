<?php
function numericValues() { yield 1; }
function arrayValues() { yield ['name' => 'a']; }
function pipeAndGeneratorOffsets() {
    $length = 'abc' |> strlen(...);
    <error descr="'$length' does not support offset access (types: int).">$length[0]</error>;
    foreach (numericValues() as $number) {
        <error descr="'$number' does not support offset access (types: int).">$number[0]</error>;
    }
    foreach (arrayValues() as $row) {
        echo $row['name'];
    }
}
