<?php
// The parent is not indexed: its load() may use static::, and subclasses
// may live elsewhere. The report stays, without a fix.
class WebformOptions extends \Drupal\Core\Config\Entity\ConfigEntityBase
{
    public static function byId($id)
    {
        return <weak_warning descr="Refer to the class as &apos;self&apos; instead of &apos;WebformOptions&apos;.">WebformOptions</weak_warning>::load($id);
    }

    // Not a call: the fix stays.
    public function copy()
    {
        return new <weak_warning descr="Refer to the class as &apos;self&apos; instead of &apos;WebformOptions&apos;.">WebformOptions</weak_warning>();
    }
}
