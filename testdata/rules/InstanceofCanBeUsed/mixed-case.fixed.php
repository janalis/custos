<?php

final class Ticket {}

function kind(object $t): array
{
    return [
        $t instanceof \Ticket,
        $t instanceof \Ticket,
    ];
}
