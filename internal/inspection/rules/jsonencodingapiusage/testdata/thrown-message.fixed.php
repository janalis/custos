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
    $text = $message . json_encode($actual, JSON_THROW_ON_ERROR);
    $error = new \RuntimeException(json_encode($actual, JSON_THROW_ON_ERROR));
    echo $text, $error->getMessage(), json_encode($actual, JSON_THROW_ON_ERROR) + 1;
    throw new \RuntimeException(code: strlen(json_encode($actual, JSON_THROW_ON_ERROR)));
}
