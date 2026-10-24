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
    public function <weak_warning descr="Declare ': string' as the return type (update the whole hierarchy with a signature refactoring).">sql</weak_warning>($x)
    {
        return 'SELECT ' . $x;
    }

    public function <weak_warning descr="Declare ': void' as the return type (update the whole hierarchy with a signature refactoring).">reset</weak_warning>()
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

    public function <weak_warning descr="Declare ': void' as the return type.">reset</weak_warning>()
    {
        echo 'reset';
    }

    public function <weak_warning descr="Declare ': void' as the return type.">close</weak_warning>()
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
    public function <weak_warning descr="Declare ': void' as the return type (update the whole hierarchy with a signature refactoring).">clear</weak_warning>();
}

final class GenericTag implements Prototype
{
    /** @return $this */
    public function <weak_warning descr="Declare ': self' as the return type.">setName</weak_warning>($name)
    {
        return $this;
    }

    public function <weak_warning descr="Declare ': void' as the return type.">clear</weak_warning>()
    {
    }
}
