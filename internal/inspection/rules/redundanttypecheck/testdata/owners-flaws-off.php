<?php
// NotOptimalIfConditions with its instanceof checks turned off: the
// checks it would report stay this rule's.
interface Sold {}
final class Ticket implements Sold {}
function chains(Ticket $t, ?Ticket $maybe): void
{
    if (<warning descr="This check is always true: the value is Ticket.">$t instanceof Ticket</warning> || $t instanceof Sold) {
        echo 1;
    }
    if ($maybe instanceof Ticket && <warning descr="This check is always false: the value is Ticket.">$maybe === null</warning>) {
        echo 2;
    }
}
