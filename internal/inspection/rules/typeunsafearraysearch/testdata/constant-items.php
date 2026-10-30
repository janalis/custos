<?php
interface Webform
{
    const PAGE = 'page';
    const MODAL = 'modal';
    const LEVEL = '10';
}

class Message
{
    const USER = 'user';
    const STATE = 'state';
    const EMPTY = '';
    const COUNT = 3;

    public function check($type, $storage)
    {
        // Constants holding plain strings compare like their literals.
        in_array($type, [Webform::PAGE, Webform::MODAL]);
        in_array($storage, [self::USER, 'custom', Message::STATE]);
        // No class extends this one to redefine them.
        in_array($storage, [static::USER, static::STATE]);
        <weak_warning descr="Pass a third argument to say whether this search must be type-strict.">in_array($type, [Webform::PAGE, Webform::LEVEL])</weak_warning>;
        <weak_warning descr="Pass a third argument to say whether this search must be type-strict.">in_array($type, [self::USER, self::EMPTY])</weak_warning>;
        <weak_warning descr="Pass a third argument to say whether this search must be type-strict.">in_array($type, [self::USER, self::COUNT])</weak_warning>;
        <weak_warning descr="Pass a third argument to say whether this search must be type-strict.">in_array($type, [self::USER, self::MISSING])</weak_warning>;
        <weak_warning descr="Pass a third argument to say whether this search must be type-strict.">in_array($type, [self::USER, $storage::STATE])</weak_warning>;
    }
}

class Base
{
    const KIND = 'base';

    public function check($kind)
    {
        // A subclass may redefine what static:: names.
        <weak_warning descr="Pass a third argument to say whether this search must be type-strict.">in_array($kind, [static::KIND])</weak_warning>;
    }
}

class Derived extends Base
{
    const KIND = '1';
}

final class Sealed extends Base
{
    const OWN = 'own';

    public function own($kind)
    {
        in_array($kind, [static::OWN, static::KIND]);
    }
}

trait Kinds
{
    public function kinds($kind)
    {
        <weak_warning descr="Pass a third argument to say whether this search must be type-strict.">in_array($kind, [static::USER])</weak_warning>;
    }
}
