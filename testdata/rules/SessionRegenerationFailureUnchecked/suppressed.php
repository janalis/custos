<?php

function signInSilenced(): void
{
    @<warning descr="Check session identifier regeneration before authenticating.">session_regenerate_id(true)</warning>;
    $_SESSION['authenticated'] = true;
}

function signInParenthesised(): void
{
    (<warning descr="Check session identifier regeneration before authenticating.">session_regenerate_id(true)</warning>);
    $_SESSION['authenticated'] = true;
}

function signInNested(): void
{
    @(@(<warning descr="Check session identifier regeneration before authenticating.">session_regenerate_id(true)</warning>));
    $_SESSION['authenticated'] = true;
}
