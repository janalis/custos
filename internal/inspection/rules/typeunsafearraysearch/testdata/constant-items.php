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
        <weak_warning descr="Pass a third argument to say whether this search must be type-strict.">in_array($storage, [static::USER, static::STATE])</weak_warning>;
        <weak_warning descr="Pass a third argument to say whether this search must be type-strict.">in_array($type, [Webform::PAGE, Webform::LEVEL])</weak_warning>;
        <weak_warning descr="Pass a third argument to say whether this search must be type-strict.">in_array($type, [self::USER, self::EMPTY])</weak_warning>;
        <weak_warning descr="Pass a third argument to say whether this search must be type-strict.">in_array($type, [self::USER, self::COUNT])</weak_warning>;
        <weak_warning descr="Pass a third argument to say whether this search must be type-strict.">in_array($type, [self::USER, self::MISSING])</weak_warning>;
        <weak_warning descr="Pass a third argument to say whether this search must be type-strict.">in_array($type, [self::USER, $storage::STATE])</weak_warning>;
    }
}
