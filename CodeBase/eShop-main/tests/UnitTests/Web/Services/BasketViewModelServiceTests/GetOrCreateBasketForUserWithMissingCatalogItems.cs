using Ardalis.Specification;
using Microsoft.eShopWeb.ApplicationCore.Entities;
using Microsoft.eShopWeb.ApplicationCore.Entities.BasketAggregate;
using Microsoft.eShopWeb.ApplicationCore.Interfaces;
using Microsoft.eShopWeb.ApplicationCore.Specifications;
using Microsoft.eShopWeb.Web.Interfaces;
using Microsoft.eShopWeb.Web.Services;
using NSubstitute;
using Xunit;

namespace Microsoft.eShopWeb.UnitTests.Web.Services.BasketViewModelServiceTests;

public class GetOrCreateBasketForUserWithMissingCatalogItems
{
    private readonly string _buyerId = "Test buyerId";
    private readonly IRepository<Basket> _mockBasketRepo = Substitute.For<IRepository<Basket>>();
    private readonly IRepository<CatalogItem> _mockItemRepo = Substitute.For<IRepository<CatalogItem>>();
    private readonly IUriComposer _mockUriComposer = Substitute.For<IUriComposer>();
    private readonly IBasketQueryService _mockBasketQueryService = Substitute.For<IBasketQueryService>();

    [Fact]
    public async Task ShouldNotThrowWhenCatalogItemIsMissing()
    {
        var basket = new Basket(_buyerId);
        basket.AddItem(1, 10.0m, 1);

        _mockBasketRepo.FirstOrDefaultAsync(Arg.Any<BasketWithItemsSpecification>(), Arg.Any<CancellationToken>())
            .Returns(basket);

        _mockItemRepo.ListAsync(Arg.Any<ISpecification<CatalogItem>>(), Arg.Any<CancellationToken>())
            .Returns(new List<CatalogItem>());

        var service = new BasketViewModelService(_mockBasketRepo, _mockItemRepo, _mockUriComposer, _mockBasketQueryService);

        var result = await service.GetOrCreateBasketForUser(_buyerId);

        Assert.NotNull(result);
        Assert.Single(result.Items);
        Assert.Null(result.Items[0].ProductName);
        Assert.Null(result.Items[0].PictureUrl);
    }

    [Fact]
    public async Task MapsExistingCatalogItemsCorrectly()
    {
        var basket = new Basket(_buyerId);
        basket.AddItem(1, 10.0m, 1);
        basket.AddItem(2, 20.0m, 2);

        var catalogItem1 = Substitute.For<CatalogItem>(1, 1, "Description 1", "Product 1", 10.0m, "pic1.jpg");
        catalogItem1.Id.Returns(1);
        var catalogItem2 = Substitute.For<CatalogItem>(2, 1, "Description 2", "Product 2", 20.0m, "pic2.jpg");
        catalogItem2.Id.Returns(2);
        var catalogItems = new List<CatalogItem> { catalogItem1, catalogItem2 };

        _mockBasketRepo.FirstOrDefaultAsync(Arg.Any<BasketWithItemsSpecification>(), Arg.Any<CancellationToken>())
            .Returns(basket);

        _mockItemRepo.ListAsync(Arg.Any<ISpecification<CatalogItem>>(), Arg.Any<CancellationToken>())
            .Returns(catalogItems);

        _mockUriComposer.ComposePicUri(Arg.Any<string>()).Returns(callInfo => $"composed_{callInfo.Arg<string>()}");

        var service = new BasketViewModelService(_mockBasketRepo, _mockItemRepo, _mockUriComposer, _mockBasketQueryService);

        var result = await service.GetOrCreateBasketForUser(_buyerId);

        Assert.NotNull(result);
        Assert.Equal(2, result.Items.Count);
        Assert.Equal("Product 1", result.Items[0].ProductName);
        Assert.Equal("Product 2", result.Items[1].ProductName);
        Assert.Equal("composed_pic1.jpg", result.Items[0].PictureUrl);
        Assert.Equal("composed_pic2.jpg", result.Items[1].PictureUrl);
    }

    [Fact]
    public async Task HandlesPartialMissingCatalogItems()
    {
        var basket = new Basket(_buyerId);
        basket.AddItem(1, 10.0m, 1);
        basket.AddItem(2, 20.0m, 2);

        var catalogItem2 = Substitute.For<CatalogItem>(2, 1, "Description 2", "Product 2", 20.0m, "pic2.jpg");
        catalogItem2.Id.Returns(2);
        var catalogItems = new List<CatalogItem> { catalogItem2 };

        _mockBasketRepo.FirstOrDefaultAsync(Arg.Any<BasketWithItemsSpecification>(), Arg.Any<CancellationToken>())
            .Returns(basket);

        _mockItemRepo.ListAsync(Arg.Any<ISpecification<CatalogItem>>(), Arg.Any<CancellationToken>())
            .Returns(catalogItems);

        _mockUriComposer.ComposePicUri(Arg.Any<string>()).Returns(callInfo => $"composed_{callInfo.Arg<string>()}");

        var service = new BasketViewModelService(_mockBasketRepo, _mockItemRepo, _mockUriComposer, _mockBasketQueryService);

        var result = await service.GetOrCreateBasketForUser(_buyerId);

        Assert.NotNull(result);
        Assert.Equal(2, result.Items.Count);

        var item1 = result.Items.Single(i => i.CatalogItemId == 1);
        Assert.Null(item1.ProductName);
        Assert.Null(item1.PictureUrl);

        var item2 = result.Items.Single(i => i.CatalogItemId == 2);
        Assert.Equal("Product 2", item2.ProductName);
        Assert.Equal("composed_pic2.jpg", item2.PictureUrl);
    }
}
