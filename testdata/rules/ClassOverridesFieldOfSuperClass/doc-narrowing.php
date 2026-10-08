<?php
namespace App;

class Model {}
class ReportModel extends Model {}

class BaseController
{
    /** @var Model|null */
    protected $model;
    /** @var string */
    protected $title;
    protected ?Model $typed = null;
    protected $plain;
}

class ReportController extends BaseController
{
    /** @var ReportModel|null */
    protected $model;
    /** @var string */
    protected <weak_warning descr="Property 'title' is already declared in \App\BaseController; drop this re-declaration.">$title</weak_warning>;
    /** @var Model|null */
    protected ?Model <weak_warning descr="Property 'typed' is already declared in \App\BaseController; drop this re-declaration.">$typed</weak_warning> = null;
    /** @var ReportModel */
    protected $plain;
}

namespace App\Anon;

function make(): object
{
    return new class extends \App\BaseController {
        /** @var \App\ReportModel|null */
        protected $model;
    };
}
