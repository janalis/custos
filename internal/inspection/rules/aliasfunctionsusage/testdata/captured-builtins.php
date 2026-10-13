<?php
namespace Stats {
    function count($items) { return 0; }

    $a = <warning descr="Use '\count(...)' instead of the alias 'sizeof(...)'.">sizeof</warning>($rows);
    $b = \<warning descr="Use 'count(...)' instead of the alias 'sizeof(...)'.">sizeof</warning>($rows);
}

namespace Imported {
    use function Stats\count;

    $c = <warning descr="Use '\count(...)' instead of the alias 'sizeof(...)'.">sizeof</warning>($rows);
    $d = <warning descr="Use 'implode(...)' instead of the alias 'join(...)'.">join</warning>(',', $rows);
}
