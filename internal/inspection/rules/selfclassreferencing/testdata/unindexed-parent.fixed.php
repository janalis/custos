<?php
// The parent is not indexed: its load() may use static::, and subclasses
// may live elsewhere. The report stays, without a fix.
class WebformOptions extends \Drupal\Core\Config\Entity\ConfigEntityBase
{
    public static function byId($id)
    {
        return WebformOptions::load($id);
    }

    // Not a call: the fix stays.
    public function copy()
    {
        return new self();
    }
}
