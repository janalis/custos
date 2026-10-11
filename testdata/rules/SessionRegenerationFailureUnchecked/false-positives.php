<?php

function signInNegated(): void
{
    // Not the silence operator: the statement is not a bare discard.
    !session_regenerate_id(true);
    $_SESSION['authenticated'] = true;
}

function signInStored(): void
{
    $ok = @session_regenerate_id(true);
    $_SESSION['authenticated'] = true;
}

function signInGuarded(): void
{
    if (!@session_regenerate_id(true)) {
        return;
    }
    $_SESSION['authenticated'] = true;
}

function signInSilencedOtherKey(): void
{
    @session_regenerate_id(true);
    $_SESSION['visits'] = true;
}
