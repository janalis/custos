<?php
// The encoded text only describes the error being thrown.
function check(mixed $actual, string $message): void
{
    if ($actual === null) {
        throw new \RuntimeException($message . ': ' . json_encode($actual));
    }
    if ($actual === false) {
        throw (new \LogicException(json_encode($actual)));
    }
    $text = $message . <weak_warning descr="Pass JSON_THROW_ON_ERROR in the flags of this call.">json_encode($actual)</weak_warning>;
    $error = new \RuntimeException(<weak_warning descr="Pass JSON_THROW_ON_ERROR in the flags of this call.">json_encode($actual)</weak_warning>);
    echo $text, $error->getMessage(), <weak_warning descr="Pass JSON_THROW_ON_ERROR in the flags of this call.">json_encode($actual)</weak_warning> + 1;
    throw new \RuntimeException(code: strlen(<weak_warning descr="Pass JSON_THROW_ON_ERROR in the flags of this call.">json_encode($actual)</weak_warning>));
}
