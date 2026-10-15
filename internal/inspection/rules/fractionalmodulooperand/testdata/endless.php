<?php
function leavesByJump($stop) {
    while (true) {
        if ($stop) { goto finished; }
    }
    finished:
    return <error descr="Use fmod for fractional operands.">9 % 0.25</error>;
}
function remainsInLoop() {
    while (1) {}
    return 9 % 0.25;
}
function decodedZeroCanEnd() {
    while ("\x30") {}
    return <error descr="Use fmod for fractional operands.">9 % 0.25</error>;
}
function internalJumpCannotExit() {
    while (true) { again: goto again; }
    return 9 % 0.25;
}
function staticNowdocCannotEnd() {
    while (<<<'TEXT'
active
TEXT) {}
    return 9 % 0.25;
}
function escapedHeredocZeroCanEnd() {
    while (<<<TEXT
\x30
TEXT) {}
    return <error descr="Use fmod for fractional operands.">9 % 0.25</error>;
}
