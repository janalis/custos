<?php
/** @return array<string, int> */
function indices(): array { return ['a' => 1]; }

function after(int $index, string $name, ?string $opt, int $n)
{
    // max() of an array of unknown values adds nothing known.
    $index = max(indices()) + 1;
    $name = preg_replace('/\W/', '_', $name);
    // A branch giving the parameter back keeps what it holds.
    $name = is_numeric($name) ? '_' . $name : $name;
    $name = $name === '' ? $name : $name;
    $opt = $opt ?: 'none';
    $name = <warning descr="Assigning a value of type int does not match the parameter's declared type.">$name ? $name : 42</warning>;
    $n = <warning descr="Assigning a value of type string does not match the parameter's declared type.">$n > 3 ? 'big' : $n</warning>;
    $n = <warning descr="Assigning a value of type string does not match the parameter's declared type.">$n > 3 ? 'big' : 'small'</warning>;
    return [$index, $name, $opt, $n];
}

interface Message {}

function errors(Message $message, array $result)
{
    // The foreach makes the parameter hold other values first.
    foreach ($result as $message) {
        $message = <warning descr="Assigning a value of type string does not match the parameter's declared type.">is_array($message) ? implode(' ', $message) : $message</warning>;
    }
}

function labels(int $label)
{
    // Each pass stores an argument first: the whole value is checked.
    foreach (func_get_args() as $label) {
        $label = <warning descr="Assigning a value of type string does not match the parameter's declared type.">$label > 1 ? 'many' : $label</warning>;
    }
}

class Chat implements Message {}

interface Response
{
    public function toArray(bool $throw = true): array;
}

function narrowed(Message $message, Response $response)
{
    if (!$message instanceof Chat) {
        throw new \LogicException('not a chat');
    }
    foreach ($response->toArray(false) as $field => $message) {
        $message = \is_array($message) ? implode(' ', $message) : $message;
    }
}

interface TokenManager
{
    /** @return string|array */
    public function replace($text);

    /** @return int|float */
    public function number($text);

    /** @return string|false */
    public function find($text);
}

function tokens(TokenManager $t, array $form, ?string $message, string $s)
{
    // A call that may return the parameter's type is not reported.
    $form = $t->replace($form);
    $message = $t->replace($message);
    $s = <warning descr="Assigning a value of type float does not match the parameter's declared type.">$t->number($s)</warning>;
    $s = $t->find($s);
    // A method not given the parameter returns what it documents.
    $form = <warning descr="Assigning a value of type string does not match the parameter's declared type.">$t->replace($s)</warning>;
    // A static method given it may return its type too.
    $form = TokenStatics::replace($form);
}

final class TokenStatics
{
    /** @return string|array */
    public static function replace($text)
    {
        return $text;
    }
}
