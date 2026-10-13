<?php

interface Exportable       {}
class Sheet implements Exportable, \Countable { public function count(): int { return 0; } }

function probe(Sheet $sheet, ?Sheet $maybe, int $id, string $name) {
    return [
        <warning descr="Prefer '$sheet instanceof \Exportable'.">in_array('Exportable', class_implements($sheet))</warning>,
        <warning descr="Prefer '$sheet instanceof \Countable'.">\in_array('Countable', \class_implements($sheet), true)</warning>,
        <warning descr="Consider '$sheet instanceof \exportable' (not an exact equivalent).">in_array('exportable', class_implements($sheet))</warning>,
        <warning descr="Consider '$maybe instanceof \Exportable' (not an exact equivalent).">in_array('Exportable', class_implements($maybe))</warning>,
        <warning descr="Consider '$id instanceof \Exportable' (not an exact equivalent).">in_array('Exportable', class_implements($id))</warning>,

        in_array('Exportable', class_parents($sheet)),
        in_array('Exportable', class_implements($name)),
        in_array('Missing\Iface', class_implements($sheet)),
    ];
}
