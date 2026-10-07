<?php
foreach ($records as $key => $record):
    <weak_warning descr="Destructure directly in the foreach header.">list($id, $label) = $record</weak_warning>;
    try {
        [$first] = $record;               // E2b: inside try
    } catch (\Throwable $t) {}
    foreach ($record as $cell) {
        [$x, $y] = $record;               // E2b: direct statement of the inner loop, which does not declare $record
        <weak_warning descr="Destructure directly in the foreach header.">[$c1, $c2] = $cell</weak_warning>;
    }
    {
        [$b1] = $record;                  // E2b: plain block
    }
    switch ($key) {
        case 1:
            [$s1] = $record;              // E2b: inside switch
    }
    while ($more) {
        [$w1] = $record;                  // E2b: inner loop
    }
endforeach;
foreach ($records as $record) <weak_warning descr="Destructure directly in the foreach header.">[$id, $label] = $record</weak_warning>;
foreach ($records as $record) if ($record) [$id, $label] = $record;   // E2b: brace-less if
