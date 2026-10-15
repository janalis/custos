<?php
enum TicketState { case Draft; case Done; }
function overwriteTicketState(TicketState &$state): void { $state = 0; }
/** @param-out TicketState $state */
function replaceTicketState(TicketState &$state): void {}

function uncertainTicket(TicketState $state)
{
    if ($state === TicketState::Draft) { return 0; }
    overwriteTicketState($state);
    // The reference output no longer has a known enum type.
    return match ($state) { TicketState::Done => 1 };
}

function knownTicket(TicketState $state)
{
    if ($state === TicketState::Draft) { return 0; }
    replaceTicketState($state);
    return <error descr="Handle every reachable enum case.">match ($state) { TicketState::Done => 1 }</error>;
}
