<?php
function probe($h, $n, $flag) {
    $r = [];
    $r[] = <warning descr="Use 'strpos($h, $n) === false' instead; it avoids building a substring.">!strstr($h, $n)</warning> == $flag;
    $r[] = (int) <warning descr="Use 'strpos($h, $n) === false' instead; it avoids building a substring.">!strstr($h, $n)</warning>;
    $r[] = !<warning descr="Use 'strpos($h, $n) === false' instead; it avoids building a substring.">!strstr($h, $n)</warning>;
    $r[] = <warning descr="Use 'stripos($h, $n) === false' instead; it avoids building a substring.">!stristr($h, $n)</warning> . 'x';
    $r[] = <warning descr="Use 'strpos($h, $n) === false' instead; it avoids building a substring.">!strstr($h, $n)</warning> + 1;
    $r[] = -<warning descr="Use 'strpos($h, $n) === false' instead; it avoids building a substring.">!strstr($h, $n)</warning>;
    $r[] = @<warning descr="Use 'strpos($h, $n) === false' instead; it avoids building a substring.">!strstr($h, $n)</warning>;
    $r[] = <warning descr="Use 'strpos($h, $n) === false' instead; it avoids building a substring.">!strstr($h, $n)</warning> <=> 1;
    $r[] = clone <warning descr="Use 'strpos($h, $n) === false' instead; it avoids building a substring.">!strstr($h, $n)</warning>;
    $r[] = <warning descr="Use 'strpos($h, $n) === false' instead; it avoids building a substring.">!strstr($h, $n)</warning> ?: 'z';
    $r[] = <warning descr="Use 'strpos($h, $n) === false' instead; it avoids building a substring.">!strstr($h, $n)</warning> ?? 'z';
    $r[] = $flag & <warning descr="Use 'strpos($h, $n) === false' instead; it avoids building a substring.">!strstr($h, $n)</warning>;
    $r[] = [<warning descr="Use 'strpos($h, $n) === false' instead; it avoids building a substring.">!strstr($h, $n)</warning>];
    $r[] = <warning descr="Use 'strpos($h, $n) !== false' instead; it avoids building a substring.">strstr($h, $n, false)</warning> ? 1 : 2;
    $r[] = <warning descr="Use 'strpos($h, $n) === false' instead; it avoids building a substring.">strstr($h, $n, true) === false</warning>;
    $r[] = <warning descr="Use 'stripos($h, $n) !== false' instead; it avoids building a substring.">FALSE !== stristr($h, $n, TRUE)</warning>;
    $r[] = <warning descr="Use 'strpos($h, $n) === false' instead; it avoids building a substring.">strstr($h, $n, \false) == false</warning>;
    return $r;
}
