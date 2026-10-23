<?php
namespace App;

use Symfony\Component\Routing\Attribute\Route;

abstract class BaseController
{
    #[Route('/admin', name: 'admin')]
    public function index(): string { return 'admin'; }

    #[\Override]
    public function label(): string { return 'base'; }
}

class DashboardController extends BaseController
{
    // Hides the inherited route: the dashboard serves no /admin page.
    public function index(): string { return parent::index(); }

}
