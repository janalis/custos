<?php
// The same chain when NotOptimalIfConditions does not run.
interface Sold {}
final class Ticket implements Sold {}
function chains(Ticket $t): void
{
    if (<warning descr="This check is always true: the value is Ticket.">$t instanceof Ticket</warning> || $t instanceof Sold) {
        echo 1;
    }
}
