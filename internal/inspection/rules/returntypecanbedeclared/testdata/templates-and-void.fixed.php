<?php
interface Reader
{
    /**
     * @template T
     * @param class-string<T> $name
     * @return T|null
     */
    public function annotation(string $name);
}

/** @template T of object */
abstract class Lexer
{
    /** @return T|null */
    abstract public function token();

    /** @return list<T> */
    abstract public function tokens();
}

abstract class Platform
{
    /** @return string */
    public function sql($x)
    {
        return 'SELECT ' . $x;
    }

    public function reset()
    {
    }
}

final class Sqlite extends Platform
{
    // The parent promises a value: no override can be void.
    public function sql($x)
    {
        throw new \LogicException('unsupported');
    }

    public function reset(): void
    {
        echo 'reset';
    }

    public function close(): void
    {
        echo 'closed';
    }
}

interface Prototype
{
    // An implementation returns a value: void cannot be declared here.
    /** @return void */
    public function setName($name);

    /** @return void */
    public function clear();
}

final class GenericTag implements Prototype
{
    /** @return $this */
    public function setName($name): self
    {
        return $this;
    }

    public function clear(): void
    {
    }
}
