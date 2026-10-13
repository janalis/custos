<?php

final class Ticket {}

function kind(object $t): array
{
    return [
        <warning descr="Prefer '$t instanceof \Ticket'.">Get_Class($t) === 'Ticket'</warning>,
        <warning descr="Prefer '$t instanceof \Ticket'.">IS_A($t, 'Ticket')</warning>,
    ];
}
