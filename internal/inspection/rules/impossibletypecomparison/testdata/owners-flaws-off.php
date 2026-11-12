<?php
// NotOptimalIfConditions with its instanceof checks turned off: the
// comparison next to an instanceof stays this rule's.
final class Ticket {}
function chains($o): void
{
    if ($o instanceof Ticket && <error descr="This comparison is always false: Ticket and string share no value.">$o === 'x'</error>) {
        echo 1;
    }
}
