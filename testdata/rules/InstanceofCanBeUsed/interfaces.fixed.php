<?php

interface Exportable       {}
class Sheet implements Exportable, \Countable { public function count(): int { return 0; } }

function probe(Sheet $sheet, ?Sheet $maybe, int $id, string $name) {
    return [
        $sheet instanceof \Exportable,
        $sheet instanceof \Countable,
        in_array('exportable', class_implements($sheet)),
        in_array('Exportable', class_implements($maybe)),
        in_array('Exportable', class_implements($id)),

        in_array('Exportable', class_parents($sheet)),
        in_array('Exportable', class_implements($name)),
        in_array('Missing\Iface', class_implements($sheet)),
    ];
}
