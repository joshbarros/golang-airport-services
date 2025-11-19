# SOLID Principles in Go

## Overview

SOLID principles help us create maintainable, flexible, and scalable code. This guide demonstrates how to apply SOLID principles in Go within the Airport Services platform.

---

## S - Single Responsibility Principle (SRP)

**Definition**: A class (or struct in Go) should have one, and only one, reason to change.

### ❌ Violation Example

```go
// BAD: FlightService has multiple responsibilities
type FlightService struct {
    db *sql.DB
}

// Database access
func (s *FlightService) GetFlight(id string) (*Flight, error) {
    // Database query logic
}

// Email sending
func (s *FlightService) SendDelayNotification(flightID string) error {
    // Email sending logic
}

// Pricing calculation
func (s *FlightService) CalculatePrice(flightID string) (float64, error) {
    // Pricing logic
}

// Report generation
func (s *FlightService) GenerateReport(flightID string) ([]byte, error) {
    // Report generation logic
}
```

**Problems:**
- Changes to email logic affect FlightService
- Changes to pricing affect FlightService
- Hard to test
- Violates separation of concerns

### ✅ Correct Example

```go
// GOOD: Each struct has a single responsibility

// 1. Flight Repository - Data access only
type FlightRepository struct {
    db *sql.DB
}

func (r *FlightRepository) FindByID(ctx context.Context, id string) (*entity.Flight, error) {
    // Only database access logic
    query := "SELECT * FROM flights WHERE id = $1"
    // ...
}

func (r *FlightRepository) Save(ctx context.Context, flight *entity.Flight) error {
    // Only database persistence logic
}

// 2. Notification Service - Communication only
type NotificationService struct {
    emailClient EmailClient
    smsClient   SMSClient
}

func (s *NotificationService) SendDelayNotification(ctx context.Context, flight *entity.Flight) error {
    // Only notification logic
}

// 3. Pricing Service - Pricing calculations only
type PricingService struct {
    ruleEngine PricingRuleEngine
}

func (s *PricingService) CalculateFlightPrice(ctx context.Context, flight *entity.Flight) (Money, error) {
    // Only pricing logic
}

// 4. Report Generator - Report generation only
type FlightReportGenerator struct {
    templateEngine TemplateEngine
}

func (g *FlightReportGenerator) GenerateReport(ctx context.Context, flight *entity.Flight) ([]byte, error) {
    // Only report generation logic
}
```

**Benefits:**
- Each component has one reason to change
- Easy to test independently
- Clear responsibility boundaries
- Better maintainability

---

## O - Open/Closed Principle (OCP)

**Definition**: Software entities should be open for extension but closed for modification.

### ❌ Violation Example

```go
// BAD: Must modify PaymentProcessor for each new payment method
type PaymentProcessor struct{}

func (p *PaymentProcessor) ProcessPayment(method string, amount Money) error {
    switch method {
    case "credit_card":
        // Credit card logic
        return p.processCreditCard(amount)
    case "paypal":
        // PayPal logic
        return p.processPayPal(amount)
    case "bank_transfer":
        // Bank transfer logic
        return p.processBankTransfer(amount)
    // Need to modify this function for every new payment method!
    default:
        return errors.New("unsupported payment method")
    }
}
```

**Problem**: Adding new payment methods requires modifying existing code.

### ✅ Correct Example

```go
// GOOD: Open for extension, closed for modification

// 1. Define interface
type PaymentMethod interface {
    ProcessPayment(ctx context.Context, amount Money) (*PaymentResult, error)
    ValidatePayment(ctx context.Context, details PaymentDetails) error
    GetMethodName() string
}

// 2. Implement for each payment method
type CreditCardPayment struct {
    gateway CreditCardGateway
}

func (c *CreditCardPayment) ProcessPayment(ctx context.Context, amount Money) (*PaymentResult, error) {
    // Credit card specific logic
    return c.gateway.Charge(ctx, amount)
}

func (c *CreditCardPayment) ValidatePayment(ctx context.Context, details PaymentDetails) error {
    // Credit card validation
    return c.validateCard(details.CardNumber, details.CVV)
}

func (c *CreditCardPayment) GetMethodName() string {
    return "credit_card"
}

type PayPalPayment struct {
    client PayPalClient
}

func (p *PayPalPayment) ProcessPayment(ctx context.Context, amount Money) (*PaymentResult, error) {
    // PayPal specific logic
    return p.client.CreatePayment(ctx, amount)
}

func (p *PayPalPayment) ValidatePayment(ctx context.Context, details PaymentDetails) error {
    // PayPal validation
    return p.client.ValidateAccount(details.PayPalEmail)
}

func (p *PayPalPayment) GetMethodName() string {
    return "paypal"
}

// 3. Payment processor doesn't need modification for new methods
type PaymentProcessor struct {
    methods map[string]PaymentMethod
}

func NewPaymentProcessor() *PaymentProcessor {
    return &PaymentProcessor{
        methods: make(map[string]PaymentMethod),
    }
}

// Register new payment methods without modifying existing code
func (p *PaymentProcessor) RegisterMethod(method PaymentMethod) {
    p.methods[method.GetMethodName()] = method
}

func (p *PaymentProcessor) ProcessPayment(ctx context.Context, methodName string, amount Money, details PaymentDetails) (*PaymentResult, error) {
    method, exists := p.methods[methodName]
    if !exists {
        return nil, ErrPaymentMethodNotSupported
    }

    if err := method.ValidatePayment(ctx, details); err != nil {
        return nil, err
    }

    return method.ProcessPayment(ctx, amount)
}

// Usage: Adding new payment method doesn't require changing existing code
func main() {
    processor := NewPaymentProcessor()

    // Register existing methods
    processor.RegisterMethod(NewCreditCardPayment(ccGateway))
    processor.RegisterMethod(NewPayPalPayment(paypalClient))

    // Add new payment method (extension, not modification!)
    processor.RegisterMethod(NewCryptoPayment(cryptoGateway))
    processor.RegisterMethod(NewApplePayPayment(applePayClient))
}
```

### Another Example: Notification Channels

```go
// Notification interface
type NotificationChannel interface {
    Send(ctx context.Context, recipient string, message string) error
    SupportsChannel() string
}

// Email channel
type EmailChannel struct {
    smtpClient SMTPClient
}

func (e *EmailChannel) Send(ctx context.Context, recipient string, message string) error {
    return e.smtpClient.SendEmail(recipient, message)
}

func (e *EmailChannel) SupportsChannel() string {
    return "email"
}

// SMS channel
type SMSChannel struct {
    smsGateway SMSGateway
}

func (s *SMSChannel) Send(ctx context.Context, recipient string, message string) error {
    return s.smsGateway.SendSMS(recipient, message)
}

func (s *SMSChannel) SupportsChannel() string {
    return "sms"
}

// Push notification channel (new, no existing code modified)
type PushChannel struct {
    pushService PushNotificationService
}

func (p *PushChannel) Send(ctx context.Context, recipient string, message string) error {
    return p.pushService.SendPush(recipient, message)
}

func (p *PushChannel) SupportsChannel() string {
    return "push"
}

// Notification service (never needs modification)
type NotificationService struct {
    channels map[string]NotificationChannel
}

func (n *NotificationService) RegisterChannel(channel NotificationChannel) {
    n.channels[channel.SupportsChannel()] = channel
}

func (n *NotificationService) SendNotification(ctx context.Context, channelType, recipient, message string) error {
    channel, exists := n.channels[channelType]
    if !exists {
        return ErrChannelNotSupported
    }
    return channel.Send(ctx, recipient, message)
}
```

---

## L - Liskov Substitution Principle (LSP)

**Definition**: Subtypes must be substitutable for their base types without altering program correctness.

### ❌ Violation Example

```go
// BAD: LSP violation

type Flight interface {
    GetDuration() time.Duration
    CalculateFuelNeeded() float64
}

// Regular flight
type RegularFlight struct {
    distance float64
}

func (f *RegularFlight) GetDuration() time.Duration {
    return time.Duration(f.distance/500) * time.Hour
}

func (f *RegularFlight) CalculateFuelNeeded() float64 {
    return f.distance * 2.5
}

// Charter flight - violates LSP!
type CharterFlight struct {
    RegularFlight
}

func (c *CharterFlight) CalculateFuelNeeded() float64 {
    // Violates LSP: throws panic instead of returning value
    panic("charter flights calculate fuel differently")
}

// This will panic when using CharterFlight!
func ProcessFlight(f Flight) {
    duration := f.GetDuration()
    fuel := f.CalculateFuelNeeded() // Panics if f is CharterFlight
    fmt.Printf("Duration: %v, Fuel: %.2f\n", duration, fuel)
}
```

### ✅ Correct Example

```go
// GOOD: Proper LSP compliance

type Flight interface {
    GetDuration() time.Duration
    GetDistance() float64
}

// Separate interface for flights that calculate fuel
type FuelCalculable interface {
    CalculateFuelNeeded() float64
}

type RegularFlight struct {
    distance      float64
    averageSpeed  float64
    fuelPerKM     float64
}

func (f *RegularFlight) GetDuration() time.Duration {
    return time.Duration(f.distance/f.averageSpeed) * time.Hour
}

func (f *RegularFlight) GetDistance() float64 {
    return f.distance
}

func (f *RegularFlight) CalculateFuelNeeded() float64 {
    return f.distance * f.fuelPerKM
}

type CharterFlight struct {
    distance     float64
    averageSpeed float64
    // Charter flights don't calculate fuel the same way
}

func (c *CharterFlight) GetDuration() time.Duration {
    return time.Duration(c.distance/c.averageSpeed) * time.Hour
}

func (c *CharterFlight) GetDistance() float64 {
    return c.distance
}

// Usage: Code works with any Flight type
func ProcessFlight(f Flight) {
    duration := f.GetDuration()
    distance := f.GetDistance()
    fmt.Printf("Duration: %v, Distance: %.2f\n", duration, distance)

    // Only calculate fuel if the flight supports it
    if fuelCalculable, ok := f.(FuelCalculable); ok {
        fuel := fuelCalculable.CalculateFuelNeeded()
        fmt.Printf("Fuel needed: %.2f\n", fuel)
    }
}
```

### Another Example: Booking Types

```go
// Base interface
type Booking interface {
    GetBookingID() string
    GetPassengerCount() int
    GetTotalPrice() Money
    Validate() error
}

// Regular booking
type StandardBooking struct {
    id         string
    passengers int
    price      Money
}

func (b *StandardBooking) GetBookingID() string {
    return b.id
}

func (b *StandardBooking) GetPassengerCount() int {
    return b.passengers
}

func (b *StandardBooking) GetTotalPrice() Money {
    return b.price
}

func (b *StandardBooking) Validate() error {
    if b.passengers == 0 {
        return errors.New("no passengers")
    }
    return nil
}

// Corporate booking - fully substitutable
type CorporateBooking struct {
    StandardBooking
    companyID   string
    billingCode string
}

// Override but maintain contract
func (c *CorporateBooking) Validate() error {
    // Call parent validation
    if err := c.StandardBooking.Validate(); err != nil {
        return err
    }

    // Additional validation (doesn't violate contract)
    if c.companyID == "" {
        return errors.New("company ID required")
    }

    return nil
}

// Works with any Booking type
func ProcessBooking(b Booking) error {
    if err := b.Validate(); err != nil {
        return err
    }

    price := b.GetTotalPrice()
    passengers := b.GetPassengerCount()

    fmt.Printf("Processing booking %s for %d passengers, total: %s\n",
        b.GetBookingID(), passengers, price)

    return nil
}
```

---

## I - Interface Segregation Principle (ISP)

**Definition**: No client should be forced to depend on methods it does not use.

### ❌ Violation Example

```go
// BAD: Fat interface forces clients to implement methods they don't need

type PassengerService interface {
    CreatePassenger(ctx context.Context, data PassengerData) error
    UpdatePassenger(ctx context.Context, id string, data PassengerData) error
    DeletePassenger(ctx context.Context, id string) error
    GetPassenger(ctx context.Context, id string) (*Passenger, error)
    SearchPassengers(ctx context.Context, query string) ([]*Passenger, error)
    ExportPassengersToCSV(ctx context.Context) ([]byte, error)
    ImportPassengersFromCSV(ctx context.Context, data []byte) error
    SendMarketingEmail(ctx context.Context, passengerID string) error
    GetPassengerLoyaltyPoints(ctx context.Context, passengerID string) (int, error)
    GeneratePassengerReport(ctx context.Context, passengerID string) ([]byte, error)
}

// CheckInService only needs GetPassenger, but must implement all methods!
type CheckInService struct {
    passengerService PassengerService
}

// Forced to implement unused methods or have compile errors
```

### ✅ Correct Example

```go
// GOOD: Small, focused interfaces

// Read operations
type PassengerReader interface {
    GetPassenger(ctx context.Context, id string) (*Passenger, error)
    SearchPassengers(ctx context.Context, query string) ([]*Passenger, error)
}

// Write operations
type PassengerWriter interface {
    CreatePassenger(ctx context.Context, data PassengerData) error
    UpdatePassenger(ctx context.Context, id string, data PassengerData) error
    DeletePassenger(ctx context.Context, id string) error
}

// Import/Export operations
type PassengerImportExport interface {
    ExportToCSV(ctx context.Context) ([]byte, error)
    ImportFromCSV(ctx context.Context, data []byte) error
}

// Marketing operations
type PassengerMarketing interface {
    SendMarketingEmail(ctx context.Context, passengerID string) error
}

// Loyalty operations
type PassengerLoyalty interface {
    GetLoyaltyPoints(ctx context.Context, passengerID string) (int, error)
}

// Reporting operations
type PassengerReporting interface {
    GenerateReport(ctx context.Context, passengerID string) ([]byte, error)
}

// Now clients only depend on what they need

// CheckInService only needs reading
type CheckInService struct {
    passengerReader PassengerReader
}

func (s *CheckInService) CheckIn(ctx context.Context, passengerID string) error {
    passenger, err := s.passengerReader.GetPassenger(ctx, passengerID)
    if err != nil {
        return err
    }
    // Check-in logic
    return nil
}

// AdminService needs both reading and writing
type AdminService struct {
    passengerReader PassengerReader
    passengerWriter PassengerWriter
}

// MarketingService only needs specific operations
type MarketingService struct {
    passengerReader    PassengerReader
    passengerMarketing PassengerMarketing
}

// Full implementation can compose all interfaces
type PassengerRepository struct {
    db *sql.DB
}

// Implement only needed interfaces
func (r *PassengerRepository) GetPassenger(ctx context.Context, id string) (*Passenger, error) {
    // Implementation
}

func (r *PassengerRepository) CreatePassenger(ctx context.Context, data PassengerData) error {
    // Implementation
}

// ... etc
```

### Flight Service Example

```go
// Small, focused interfaces

type FlightReader interface {
    GetFlight(ctx context.Context, id string) (*Flight, error)
    SearchFlights(ctx context.Context, criteria SearchCriteria) ([]*Flight, error)
}

type FlightStatusUpdater interface {
    UpdateStatus(ctx context.Context, flightID string, status FlightStatus) error
}

type FlightScheduler interface {
    ScheduleFlight(ctx context.Context, flight *Flight) error
    RescheduleFlight(ctx context.Context, flightID string, newTime time.Time) error
}

type FlightCanceller interface {
    CancelFlight(ctx context.Context, flightID string, reason string) error
}

// Display board only needs reading
type DepartureBoardService struct {
    flightReader FlightReader
}

// Air traffic control needs status updates
type AirTrafficControlService struct {
    flightStatusUpdater FlightStatusUpdater
}

// Operations team needs scheduling
type OperationsService struct {
    flightScheduler FlightScheduler
    flightCanceller FlightCanceller
}

// Full service composes all capabilities
type FlightService struct {
    repo FlightRepository // implements all interfaces
}

func (s *FlightService) GetFlight(ctx context.Context, id string) (*Flight, error) {
    return s.repo.GetFlight(ctx, id)
}

func (s *FlightService) UpdateStatus(ctx context.Context, flightID string, status FlightStatus) error {
    return s.repo.UpdateStatus(ctx, flightID, status)
}

// ... etc
```

---

## D - Dependency Inversion Principle (DIP)

**Definition**: High-level modules should not depend on low-level modules. Both should depend on abstractions.

### ❌ Violation Example

```go
// BAD: High-level use case depends on low-level implementation

// Low-level module (concrete implementation)
type PostgresFlightRepository struct {
    db *sql.DB
}

func (r *PostgresFlightRepository) SaveFlight(flight *Flight) error {
    query := "INSERT INTO flights..."
    // PostgreSQL specific code
}

// High-level module depends directly on low-level module!
type CreateFlightUseCase struct {
    repo *PostgresFlightRepository // Concrete dependency!
}

func NewCreateFlightUseCase(db *sql.DB) *CreateFlightUseCase {
    return &CreateFlightUseCase{
        repo: &PostgresFlightRepository{db: db}, // Tight coupling!
    }
}

func (uc *CreateFlightUseCase) Execute(ctx context.Context, input Input) error {
    flight := createFlight(input)
    return uc.repo.SaveFlight(flight) // Depends on concrete type
}
```

**Problems:**
- Cannot easily switch to MongoDB or other database
- Hard to test (requires real PostgreSQL)
- High-level logic coupled to low-level details

### ✅ Correct Example

```go
// GOOD: Both depend on abstraction

// 1. Define abstraction (interface) in domain layer
package repository

type FlightRepository interface {
    Save(ctx context.Context, flight *entity.Flight) error
    FindByID(ctx context.Context, id string) (*entity.Flight, error)
    Delete(ctx context.Context, id string) error
}

// 2. High-level module depends on interface
package usecase

type CreateFlightUseCase struct {
    repo repository.FlightRepository // Abstract dependency!
}

// Depend on interface, not concrete type
func NewCreateFlightUseCase(repo repository.FlightRepository) *CreateFlightUseCase {
    return &CreateFlightUseCase{
        repo: repo,
    }
}

func (uc *CreateFlightUseCase) Execute(ctx context.Context, input Input) error {
    flight := entity.NewFlight(input)

    // Works with any implementation of FlightRepository
    return uc.repo.Save(ctx, flight)
}

// 3. Low-level implementation depends on same interface
package postgres

type PostgresFlightRepository struct {
    db *sql.DB
}

// Implements the interface
func (r *PostgresFlightRepository) Save(ctx context.Context, flight *entity.Flight) error {
    // PostgreSQL implementation
}

func (r *PostgresFlightRepository) FindByID(ctx context.Context, id string) (*entity.Flight, error) {
    // PostgreSQL implementation
}

// 4. Can have multiple implementations
package mongodb

type MongoFlightRepository struct {
    collection *mongo.Collection
}

// Also implements the same interface
func (r *MongoFlightRepository) Save(ctx context.Context, flight *entity.Flight) error {
    // MongoDB implementation
}

func (r *MongoFlightRepository) FindByID(ctx context.Context, id string) (*entity.Flight, error) {
    // MongoDB implementation
}

// 5. Dependency injection at composition root
func main() {
    // Can easily swap implementations
    db := setupPostgresDB()
    flightRepo := postgres.NewPostgresFlightRepository(db)

    // Or use MongoDB
    // mongoClient := setupMongoDB()
    // flightRepo := mongodb.NewMongoFlightRepository(mongoClient)

    // Use case doesn't care which implementation
    createFlightUC := usecase.NewCreateFlightUseCase(flightRepo)
}
```

### Complete Example with Multiple Dependencies

```go
// Domain layer interfaces (abstractions)

package repository

type BookingRepository interface {
    Save(ctx context.Context, booking *entity.Booking) error
    FindByID(ctx context.Context, id string) (*entity.Booking, error)
}

type PaymentGateway interface {
    ProcessPayment(ctx context.Context, amount Money, method PaymentMethod) (*PaymentResult, error)
    RefundPayment(ctx context.Context, transactionID string) error
}

type NotificationSender interface {
    SendEmail(ctx context.Context, to, subject, body string) error
    SendSMS(ctx context.Context, to, message string) error
}

// High-level use case depends on abstractions
package usecase

type CreateBookingUseCase struct {
    bookingRepo  repository.BookingRepository   // Interface
    paymentGW    repository.PaymentGateway      // Interface
    notifier     repository.NotificationSender  // Interface
}

func NewCreateBookingUseCase(
    bookingRepo repository.BookingRepository,
    paymentGW repository.PaymentGateway,
    notifier repository.NotificationSender,
) *CreateBookingUseCase {
    return &CreateBookingUseCase{
        bookingRepo: bookingRepo,
        paymentGW:   paymentGW,
        notifier:    notifier,
    }
}

func (uc *CreateBookingUseCase) Execute(ctx context.Context, input Input) error {
    // Create booking
    booking := entity.NewBooking(input)

    // Process payment (doesn't care if Stripe, PayPal, etc.)
    payment, err := uc.paymentGW.ProcessPayment(ctx, booking.TotalAmount(), input.PaymentMethod)
    if err != nil {
        return err
    }

    booking.MarkAsPaid(payment.TransactionID)

    // Save booking (doesn't care if Postgres, MongoDB, etc.)
    if err := uc.bookingRepo.Save(ctx, booking); err != nil {
        // Refund if save fails
        uc.paymentGW.RefundPayment(ctx, payment.TransactionID)
        return err
    }

    // Send confirmation (doesn't care if SMTP, SendGrid, etc.)
    uc.notifier.SendEmail(ctx, booking.PassengerEmail(), "Booking Confirmed", "Your booking is confirmed")

    return nil
}

// Low-level implementations

package postgres

type PostgresBookingRepository struct {
    db *sql.DB
}

func (r *PostgresBookingRepository) Save(ctx context.Context, booking *entity.Booking) error {
    // Postgres implementation
}

package stripe

type StripePaymentGateway struct {
    client *stripe.Client
}

func (g *StripePaymentGateway) ProcessPayment(ctx context.Context, amount Money, method PaymentMethod) (*PaymentResult, error) {
    // Stripe implementation
}

package sendgrid

type SendGridNotifier struct {
    client *sendgrid.Client
}

func (n *SendGridNotifier) SendEmail(ctx context.Context, to, subject, body string) error {
    // SendGrid implementation
}

// Composition root (main.go)
func main() {
    // Initialize implementations
    db := setupPostgresDB()
    bookingRepo := postgres.NewPostgresBookingRepository(db)

    stripeClient := stripe.NewClient(config.StripeAPIKey)
    paymentGW := stripe.NewStripePaymentGateway(stripeClient)

    sendgridClient := sendgrid.NewClient(config.SendGridAPIKey)
    notifier := sendgrid.NewSendGridNotifier(sendgridClient)

    // Inject dependencies
    createBookingUC := usecase.NewCreateBookingUseCase(
        bookingRepo,
        paymentGW,
        notifier,
    )

    // Use case works with any implementations!
}
```

### Testing with DIP

```go
// Easy to test with mock implementations

type MockBookingRepository struct {
    mock.Mock
}

func (m *MockBookingRepository) Save(ctx context.Context, booking *entity.Booking) error {
    args := m.Called(ctx, booking)
    return args.Error(0)
}

func (m *MockBookingRepository) FindByID(ctx context.Context, id string) (*entity.Booking, error) {
    args := m.Called(ctx, id)
    return args.Get(0).(*entity.Booking), args.Error(1)
}

func TestCreateBookingUseCase_Execute(t *testing.T) {
    // Arrange
    mockRepo := new(MockBookingRepository)
    mockPayment := new(MockPaymentGateway)
    mockNotifier := new(MockNotificationSender)

    uc := usecase.NewCreateBookingUseCase(mockRepo, mockPayment, mockNotifier)

    mockRepo.On("Save", mock.Anything, mock.Anything).Return(nil)
    mockPayment.On("ProcessPayment", mock.Anything, mock.Anything, mock.Anything).
        Return(&PaymentResult{TransactionID: "123"}, nil)
    mockNotifier.On("SendEmail", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
        Return(nil)

    input := createTestInput()

    // Act
    err := uc.Execute(context.Background(), input)

    // Assert
    assert.NoError(t, err)
    mockRepo.AssertExpectations(t)
    mockPayment.AssertExpectations(t)
    mockNotifier.AssertExpectations(t)
}
```

---

## Summary

### Single Responsibility Principle (SRP)
- ✅ One struct, one responsibility
- ✅ Separate concerns into different types
- ✅ Easy to understand and maintain

### Open/Closed Principle (OCP)
- ✅ Use interfaces for extension points
- ✅ Add new features without modifying existing code
- ✅ Strategy pattern, plugin architectures

### Liskov Substitution Principle (LSP)
- ✅ Subtypes must honor base type contracts
- ✅ Don't throw unexpected errors
- ✅ Maintain expected behavior

### Interface Segregation Principle (ISP)
- ✅ Many small interfaces > one large interface
- ✅ Clients depend only on methods they use
- ✅ Reduces coupling

### Dependency Inversion Principle (DIP)
- ✅ Depend on interfaces, not concrete types
- ✅ High-level modules independent of low-level details
- ✅ Easy to test and swap implementations

## Benefits of SOLID in Airport Services

1. **Testability**: Mock implementations for unit testing
2. **Flexibility**: Swap databases, payment gateways without changing business logic
3. **Maintainability**: Clear responsibilities, easy to locate code
4. **Scalability**: Add new features without breaking existing code
5. **Team Collaboration**: Clear boundaries enable parallel development

By following SOLID principles, we create a codebase that is robust, maintainable, and ready for growth.
