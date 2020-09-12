// Code generated by SQLBoiler 4.2.0 (https://github.com/volatiletech/sqlboiler). DO NOT EDIT.
// This file is meant to be re-generated in place and/or deleted at any time.

package models

import (
	"bytes"
	"context"
	"reflect"
	"testing"

	"github.com/volatiletech/randomize"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries"
	"github.com/volatiletech/strmangle"
)

var (
	// Relationships sometimes use the reflection helper queries.Equal/queries.Assign
	// so force a package dependency in case they don't.
	_ = queries.Equal
)

func testPaymentStatuses(t *testing.T) {
	t.Parallel()

	query := PaymentStatuses()

	if query.Query == nil {
		t.Error("expected a query, got nothing")
	}
}

func testPaymentStatusesDelete(t *testing.T) {
	t.Parallel()

	seed := randomize.NewSeed()
	var err error
	o := &PaymentStatus{}
	if err = randomize.Struct(seed, o, paymentStatusDBTypes, true, paymentStatusColumnsWithDefault...); err != nil {
		t.Errorf("Unable to randomize PaymentStatus struct: %s", err)
	}

	ctx := context.Background()
	tx := MustTx(boil.BeginTx(ctx, nil))
	defer func() { _ = tx.Rollback() }()
	if err = o.Insert(ctx, tx, boil.Infer()); err != nil {
		t.Error(err)
	}

	if rowsAff, err := o.Delete(ctx, tx); err != nil {
		t.Error(err)
	} else if rowsAff != 1 {
		t.Error("should only have deleted one row, but affected:", rowsAff)
	}

	count, err := PaymentStatuses().Count(ctx, tx)
	if err != nil {
		t.Error(err)
	}

	if count != 0 {
		t.Error("want zero records, got:", count)
	}
}

func testPaymentStatusesQueryDeleteAll(t *testing.T) {
	t.Parallel()

	seed := randomize.NewSeed()
	var err error
	o := &PaymentStatus{}
	if err = randomize.Struct(seed, o, paymentStatusDBTypes, true, paymentStatusColumnsWithDefault...); err != nil {
		t.Errorf("Unable to randomize PaymentStatus struct: %s", err)
	}

	ctx := context.Background()
	tx := MustTx(boil.BeginTx(ctx, nil))
	defer func() { _ = tx.Rollback() }()
	if err = o.Insert(ctx, tx, boil.Infer()); err != nil {
		t.Error(err)
	}

	if rowsAff, err := PaymentStatuses().DeleteAll(ctx, tx); err != nil {
		t.Error(err)
	} else if rowsAff != 1 {
		t.Error("should only have deleted one row, but affected:", rowsAff)
	}

	count, err := PaymentStatuses().Count(ctx, tx)
	if err != nil {
		t.Error(err)
	}

	if count != 0 {
		t.Error("want zero records, got:", count)
	}
}

func testPaymentStatusesSliceDeleteAll(t *testing.T) {
	t.Parallel()

	seed := randomize.NewSeed()
	var err error
	o := &PaymentStatus{}
	if err = randomize.Struct(seed, o, paymentStatusDBTypes, true, paymentStatusColumnsWithDefault...); err != nil {
		t.Errorf("Unable to randomize PaymentStatus struct: %s", err)
	}

	ctx := context.Background()
	tx := MustTx(boil.BeginTx(ctx, nil))
	defer func() { _ = tx.Rollback() }()
	if err = o.Insert(ctx, tx, boil.Infer()); err != nil {
		t.Error(err)
	}

	slice := PaymentStatusSlice{o}

	if rowsAff, err := slice.DeleteAll(ctx, tx); err != nil {
		t.Error(err)
	} else if rowsAff != 1 {
		t.Error("should only have deleted one row, but affected:", rowsAff)
	}

	count, err := PaymentStatuses().Count(ctx, tx)
	if err != nil {
		t.Error(err)
	}

	if count != 0 {
		t.Error("want zero records, got:", count)
	}
}

func testPaymentStatusesExists(t *testing.T) {
	t.Parallel()

	seed := randomize.NewSeed()
	var err error
	o := &PaymentStatus{}
	if err = randomize.Struct(seed, o, paymentStatusDBTypes, true, paymentStatusColumnsWithDefault...); err != nil {
		t.Errorf("Unable to randomize PaymentStatus struct: %s", err)
	}

	ctx := context.Background()
	tx := MustTx(boil.BeginTx(ctx, nil))
	defer func() { _ = tx.Rollback() }()
	if err = o.Insert(ctx, tx, boil.Infer()); err != nil {
		t.Error(err)
	}

	e, err := PaymentStatusExists(ctx, tx, o.ID)
	if err != nil {
		t.Errorf("Unable to check if PaymentStatus exists: %s", err)
	}
	if !e {
		t.Errorf("Expected PaymentStatusExists to return true, but got false.")
	}
}

func testPaymentStatusesFind(t *testing.T) {
	t.Parallel()

	seed := randomize.NewSeed()
	var err error
	o := &PaymentStatus{}
	if err = randomize.Struct(seed, o, paymentStatusDBTypes, true, paymentStatusColumnsWithDefault...); err != nil {
		t.Errorf("Unable to randomize PaymentStatus struct: %s", err)
	}

	ctx := context.Background()
	tx := MustTx(boil.BeginTx(ctx, nil))
	defer func() { _ = tx.Rollback() }()
	if err = o.Insert(ctx, tx, boil.Infer()); err != nil {
		t.Error(err)
	}

	paymentStatusFound, err := FindPaymentStatus(ctx, tx, o.ID)
	if err != nil {
		t.Error(err)
	}

	if paymentStatusFound == nil {
		t.Error("want a record, got nil")
	}
}

func testPaymentStatusesBind(t *testing.T) {
	t.Parallel()

	seed := randomize.NewSeed()
	var err error
	o := &PaymentStatus{}
	if err = randomize.Struct(seed, o, paymentStatusDBTypes, true, paymentStatusColumnsWithDefault...); err != nil {
		t.Errorf("Unable to randomize PaymentStatus struct: %s", err)
	}

	ctx := context.Background()
	tx := MustTx(boil.BeginTx(ctx, nil))
	defer func() { _ = tx.Rollback() }()
	if err = o.Insert(ctx, tx, boil.Infer()); err != nil {
		t.Error(err)
	}

	if err = PaymentStatuses().Bind(ctx, tx, o); err != nil {
		t.Error(err)
	}
}

func testPaymentStatusesOne(t *testing.T) {
	t.Parallel()

	seed := randomize.NewSeed()
	var err error
	o := &PaymentStatus{}
	if err = randomize.Struct(seed, o, paymentStatusDBTypes, true, paymentStatusColumnsWithDefault...); err != nil {
		t.Errorf("Unable to randomize PaymentStatus struct: %s", err)
	}

	ctx := context.Background()
	tx := MustTx(boil.BeginTx(ctx, nil))
	defer func() { _ = tx.Rollback() }()
	if err = o.Insert(ctx, tx, boil.Infer()); err != nil {
		t.Error(err)
	}

	if x, err := PaymentStatuses().One(ctx, tx); err != nil {
		t.Error(err)
	} else if x == nil {
		t.Error("expected to get a non nil record")
	}
}

func testPaymentStatusesAll(t *testing.T) {
	t.Parallel()

	seed := randomize.NewSeed()
	var err error
	paymentStatusOne := &PaymentStatus{}
	paymentStatusTwo := &PaymentStatus{}
	if err = randomize.Struct(seed, paymentStatusOne, paymentStatusDBTypes, false, paymentStatusColumnsWithDefault...); err != nil {
		t.Errorf("Unable to randomize PaymentStatus struct: %s", err)
	}
	if err = randomize.Struct(seed, paymentStatusTwo, paymentStatusDBTypes, false, paymentStatusColumnsWithDefault...); err != nil {
		t.Errorf("Unable to randomize PaymentStatus struct: %s", err)
	}

	ctx := context.Background()
	tx := MustTx(boil.BeginTx(ctx, nil))
	defer func() { _ = tx.Rollback() }()
	if err = paymentStatusOne.Insert(ctx, tx, boil.Infer()); err != nil {
		t.Error(err)
	}
	if err = paymentStatusTwo.Insert(ctx, tx, boil.Infer()); err != nil {
		t.Error(err)
	}

	slice, err := PaymentStatuses().All(ctx, tx)
	if err != nil {
		t.Error(err)
	}

	if len(slice) != 2 {
		t.Error("want 2 records, got:", len(slice))
	}
}

func testPaymentStatusesCount(t *testing.T) {
	t.Parallel()

	var err error
	seed := randomize.NewSeed()
	paymentStatusOne := &PaymentStatus{}
	paymentStatusTwo := &PaymentStatus{}
	if err = randomize.Struct(seed, paymentStatusOne, paymentStatusDBTypes, false, paymentStatusColumnsWithDefault...); err != nil {
		t.Errorf("Unable to randomize PaymentStatus struct: %s", err)
	}
	if err = randomize.Struct(seed, paymentStatusTwo, paymentStatusDBTypes, false, paymentStatusColumnsWithDefault...); err != nil {
		t.Errorf("Unable to randomize PaymentStatus struct: %s", err)
	}

	ctx := context.Background()
	tx := MustTx(boil.BeginTx(ctx, nil))
	defer func() { _ = tx.Rollback() }()
	if err = paymentStatusOne.Insert(ctx, tx, boil.Infer()); err != nil {
		t.Error(err)
	}
	if err = paymentStatusTwo.Insert(ctx, tx, boil.Infer()); err != nil {
		t.Error(err)
	}

	count, err := PaymentStatuses().Count(ctx, tx)
	if err != nil {
		t.Error(err)
	}

	if count != 2 {
		t.Error("want 2 records, got:", count)
	}
}

func paymentStatusBeforeInsertHook(ctx context.Context, e boil.ContextExecutor, o *PaymentStatus) error {
	*o = PaymentStatus{}
	return nil
}

func paymentStatusAfterInsertHook(ctx context.Context, e boil.ContextExecutor, o *PaymentStatus) error {
	*o = PaymentStatus{}
	return nil
}

func paymentStatusAfterSelectHook(ctx context.Context, e boil.ContextExecutor, o *PaymentStatus) error {
	*o = PaymentStatus{}
	return nil
}

func paymentStatusBeforeUpdateHook(ctx context.Context, e boil.ContextExecutor, o *PaymentStatus) error {
	*o = PaymentStatus{}
	return nil
}

func paymentStatusAfterUpdateHook(ctx context.Context, e boil.ContextExecutor, o *PaymentStatus) error {
	*o = PaymentStatus{}
	return nil
}

func paymentStatusBeforeDeleteHook(ctx context.Context, e boil.ContextExecutor, o *PaymentStatus) error {
	*o = PaymentStatus{}
	return nil
}

func paymentStatusAfterDeleteHook(ctx context.Context, e boil.ContextExecutor, o *PaymentStatus) error {
	*o = PaymentStatus{}
	return nil
}

func paymentStatusBeforeUpsertHook(ctx context.Context, e boil.ContextExecutor, o *PaymentStatus) error {
	*o = PaymentStatus{}
	return nil
}

func paymentStatusAfterUpsertHook(ctx context.Context, e boil.ContextExecutor, o *PaymentStatus) error {
	*o = PaymentStatus{}
	return nil
}

func testPaymentStatusesHooks(t *testing.T) {
	t.Parallel()

	var err error

	ctx := context.Background()
	empty := &PaymentStatus{}
	o := &PaymentStatus{}

	seed := randomize.NewSeed()
	if err = randomize.Struct(seed, o, paymentStatusDBTypes, false); err != nil {
		t.Errorf("Unable to randomize PaymentStatus object: %s", err)
	}

	AddPaymentStatusHook(boil.BeforeInsertHook, paymentStatusBeforeInsertHook)
	if err = o.doBeforeInsertHooks(ctx, nil); err != nil {
		t.Errorf("Unable to execute doBeforeInsertHooks: %s", err)
	}
	if !reflect.DeepEqual(o, empty) {
		t.Errorf("Expected BeforeInsertHook function to empty object, but got: %#v", o)
	}
	paymentStatusBeforeInsertHooks = []PaymentStatusHook{}

	AddPaymentStatusHook(boil.AfterInsertHook, paymentStatusAfterInsertHook)
	if err = o.doAfterInsertHooks(ctx, nil); err != nil {
		t.Errorf("Unable to execute doAfterInsertHooks: %s", err)
	}
	if !reflect.DeepEqual(o, empty) {
		t.Errorf("Expected AfterInsertHook function to empty object, but got: %#v", o)
	}
	paymentStatusAfterInsertHooks = []PaymentStatusHook{}

	AddPaymentStatusHook(boil.AfterSelectHook, paymentStatusAfterSelectHook)
	if err = o.doAfterSelectHooks(ctx, nil); err != nil {
		t.Errorf("Unable to execute doAfterSelectHooks: %s", err)
	}
	if !reflect.DeepEqual(o, empty) {
		t.Errorf("Expected AfterSelectHook function to empty object, but got: %#v", o)
	}
	paymentStatusAfterSelectHooks = []PaymentStatusHook{}

	AddPaymentStatusHook(boil.BeforeUpdateHook, paymentStatusBeforeUpdateHook)
	if err = o.doBeforeUpdateHooks(ctx, nil); err != nil {
		t.Errorf("Unable to execute doBeforeUpdateHooks: %s", err)
	}
	if !reflect.DeepEqual(o, empty) {
		t.Errorf("Expected BeforeUpdateHook function to empty object, but got: %#v", o)
	}
	paymentStatusBeforeUpdateHooks = []PaymentStatusHook{}

	AddPaymentStatusHook(boil.AfterUpdateHook, paymentStatusAfterUpdateHook)
	if err = o.doAfterUpdateHooks(ctx, nil); err != nil {
		t.Errorf("Unable to execute doAfterUpdateHooks: %s", err)
	}
	if !reflect.DeepEqual(o, empty) {
		t.Errorf("Expected AfterUpdateHook function to empty object, but got: %#v", o)
	}
	paymentStatusAfterUpdateHooks = []PaymentStatusHook{}

	AddPaymentStatusHook(boil.BeforeDeleteHook, paymentStatusBeforeDeleteHook)
	if err = o.doBeforeDeleteHooks(ctx, nil); err != nil {
		t.Errorf("Unable to execute doBeforeDeleteHooks: %s", err)
	}
	if !reflect.DeepEqual(o, empty) {
		t.Errorf("Expected BeforeDeleteHook function to empty object, but got: %#v", o)
	}
	paymentStatusBeforeDeleteHooks = []PaymentStatusHook{}

	AddPaymentStatusHook(boil.AfterDeleteHook, paymentStatusAfterDeleteHook)
	if err = o.doAfterDeleteHooks(ctx, nil); err != nil {
		t.Errorf("Unable to execute doAfterDeleteHooks: %s", err)
	}
	if !reflect.DeepEqual(o, empty) {
		t.Errorf("Expected AfterDeleteHook function to empty object, but got: %#v", o)
	}
	paymentStatusAfterDeleteHooks = []PaymentStatusHook{}

	AddPaymentStatusHook(boil.BeforeUpsertHook, paymentStatusBeforeUpsertHook)
	if err = o.doBeforeUpsertHooks(ctx, nil); err != nil {
		t.Errorf("Unable to execute doBeforeUpsertHooks: %s", err)
	}
	if !reflect.DeepEqual(o, empty) {
		t.Errorf("Expected BeforeUpsertHook function to empty object, but got: %#v", o)
	}
	paymentStatusBeforeUpsertHooks = []PaymentStatusHook{}

	AddPaymentStatusHook(boil.AfterUpsertHook, paymentStatusAfterUpsertHook)
	if err = o.doAfterUpsertHooks(ctx, nil); err != nil {
		t.Errorf("Unable to execute doAfterUpsertHooks: %s", err)
	}
	if !reflect.DeepEqual(o, empty) {
		t.Errorf("Expected AfterUpsertHook function to empty object, but got: %#v", o)
	}
	paymentStatusAfterUpsertHooks = []PaymentStatusHook{}
}

func testPaymentStatusesInsert(t *testing.T) {
	t.Parallel()

	seed := randomize.NewSeed()
	var err error
	o := &PaymentStatus{}
	if err = randomize.Struct(seed, o, paymentStatusDBTypes, true, paymentStatusColumnsWithDefault...); err != nil {
		t.Errorf("Unable to randomize PaymentStatus struct: %s", err)
	}

	ctx := context.Background()
	tx := MustTx(boil.BeginTx(ctx, nil))
	defer func() { _ = tx.Rollback() }()
	if err = o.Insert(ctx, tx, boil.Infer()); err != nil {
		t.Error(err)
	}

	count, err := PaymentStatuses().Count(ctx, tx)
	if err != nil {
		t.Error(err)
	}

	if count != 1 {
		t.Error("want one record, got:", count)
	}
}

func testPaymentStatusesInsertWhitelist(t *testing.T) {
	t.Parallel()

	seed := randomize.NewSeed()
	var err error
	o := &PaymentStatus{}
	if err = randomize.Struct(seed, o, paymentStatusDBTypes, true); err != nil {
		t.Errorf("Unable to randomize PaymentStatus struct: %s", err)
	}

	ctx := context.Background()
	tx := MustTx(boil.BeginTx(ctx, nil))
	defer func() { _ = tx.Rollback() }()
	if err = o.Insert(ctx, tx, boil.Whitelist(paymentStatusColumnsWithoutDefault...)); err != nil {
		t.Error(err)
	}

	count, err := PaymentStatuses().Count(ctx, tx)
	if err != nil {
		t.Error(err)
	}

	if count != 1 {
		t.Error("want one record, got:", count)
	}
}

func testPaymentStatusesReload(t *testing.T) {
	t.Parallel()

	seed := randomize.NewSeed()
	var err error
	o := &PaymentStatus{}
	if err = randomize.Struct(seed, o, paymentStatusDBTypes, true, paymentStatusColumnsWithDefault...); err != nil {
		t.Errorf("Unable to randomize PaymentStatus struct: %s", err)
	}

	ctx := context.Background()
	tx := MustTx(boil.BeginTx(ctx, nil))
	defer func() { _ = tx.Rollback() }()
	if err = o.Insert(ctx, tx, boil.Infer()); err != nil {
		t.Error(err)
	}

	if err = o.Reload(ctx, tx); err != nil {
		t.Error(err)
	}
}

func testPaymentStatusesReloadAll(t *testing.T) {
	t.Parallel()

	seed := randomize.NewSeed()
	var err error
	o := &PaymentStatus{}
	if err = randomize.Struct(seed, o, paymentStatusDBTypes, true, paymentStatusColumnsWithDefault...); err != nil {
		t.Errorf("Unable to randomize PaymentStatus struct: %s", err)
	}

	ctx := context.Background()
	tx := MustTx(boil.BeginTx(ctx, nil))
	defer func() { _ = tx.Rollback() }()
	if err = o.Insert(ctx, tx, boil.Infer()); err != nil {
		t.Error(err)
	}

	slice := PaymentStatusSlice{o}

	if err = slice.ReloadAll(ctx, tx); err != nil {
		t.Error(err)
	}
}

func testPaymentStatusesSelect(t *testing.T) {
	t.Parallel()

	seed := randomize.NewSeed()
	var err error
	o := &PaymentStatus{}
	if err = randomize.Struct(seed, o, paymentStatusDBTypes, true, paymentStatusColumnsWithDefault...); err != nil {
		t.Errorf("Unable to randomize PaymentStatus struct: %s", err)
	}

	ctx := context.Background()
	tx := MustTx(boil.BeginTx(ctx, nil))
	defer func() { _ = tx.Rollback() }()
	if err = o.Insert(ctx, tx, boil.Infer()); err != nil {
		t.Error(err)
	}

	slice, err := PaymentStatuses().All(ctx, tx)
	if err != nil {
		t.Error(err)
	}

	if len(slice) != 1 {
		t.Error("want one record, got:", len(slice))
	}
}

var (
	paymentStatusDBTypes = map[string]string{`ID`: `integer`, `OrderID`: `integer`, `ConfirmationXML`: `xml`, `Status`: `text`, `CreatedAt`: `timestamp with time zone`}
	_                    = bytes.MinRead
)

func testPaymentStatusesUpdate(t *testing.T) {
	t.Parallel()

	if 0 == len(paymentStatusPrimaryKeyColumns) {
		t.Skip("Skipping table with no primary key columns")
	}
	if len(paymentStatusAllColumns) == len(paymentStatusPrimaryKeyColumns) {
		t.Skip("Skipping table with only primary key columns")
	}

	seed := randomize.NewSeed()
	var err error
	o := &PaymentStatus{}
	if err = randomize.Struct(seed, o, paymentStatusDBTypes, true, paymentStatusColumnsWithDefault...); err != nil {
		t.Errorf("Unable to randomize PaymentStatus struct: %s", err)
	}

	ctx := context.Background()
	tx := MustTx(boil.BeginTx(ctx, nil))
	defer func() { _ = tx.Rollback() }()
	if err = o.Insert(ctx, tx, boil.Infer()); err != nil {
		t.Error(err)
	}

	count, err := PaymentStatuses().Count(ctx, tx)
	if err != nil {
		t.Error(err)
	}

	if count != 1 {
		t.Error("want one record, got:", count)
	}

	if err = randomize.Struct(seed, o, paymentStatusDBTypes, true, paymentStatusPrimaryKeyColumns...); err != nil {
		t.Errorf("Unable to randomize PaymentStatus struct: %s", err)
	}

	if rowsAff, err := o.Update(ctx, tx, boil.Infer()); err != nil {
		t.Error(err)
	} else if rowsAff != 1 {
		t.Error("should only affect one row but affected", rowsAff)
	}
}

func testPaymentStatusesSliceUpdateAll(t *testing.T) {
	t.Parallel()

	if len(paymentStatusAllColumns) == len(paymentStatusPrimaryKeyColumns) {
		t.Skip("Skipping table with only primary key columns")
	}

	seed := randomize.NewSeed()
	var err error
	o := &PaymentStatus{}
	if err = randomize.Struct(seed, o, paymentStatusDBTypes, true, paymentStatusColumnsWithDefault...); err != nil {
		t.Errorf("Unable to randomize PaymentStatus struct: %s", err)
	}

	ctx := context.Background()
	tx := MustTx(boil.BeginTx(ctx, nil))
	defer func() { _ = tx.Rollback() }()
	if err = o.Insert(ctx, tx, boil.Infer()); err != nil {
		t.Error(err)
	}

	count, err := PaymentStatuses().Count(ctx, tx)
	if err != nil {
		t.Error(err)
	}

	if count != 1 {
		t.Error("want one record, got:", count)
	}

	if err = randomize.Struct(seed, o, paymentStatusDBTypes, true, paymentStatusPrimaryKeyColumns...); err != nil {
		t.Errorf("Unable to randomize PaymentStatus struct: %s", err)
	}

	// Remove Primary keys and unique columns from what we plan to update
	var fields []string
	if strmangle.StringSliceMatch(paymentStatusAllColumns, paymentStatusPrimaryKeyColumns) {
		fields = paymentStatusAllColumns
	} else {
		fields = strmangle.SetComplement(
			paymentStatusAllColumns,
			paymentStatusPrimaryKeyColumns,
		)
	}

	value := reflect.Indirect(reflect.ValueOf(o))
	typ := reflect.TypeOf(o).Elem()
	n := typ.NumField()

	updateMap := M{}
	for _, col := range fields {
		for i := 0; i < n; i++ {
			f := typ.Field(i)
			if f.Tag.Get("boil") == col {
				updateMap[col] = value.Field(i).Interface()
			}
		}
	}

	slice := PaymentStatusSlice{o}
	if rowsAff, err := slice.UpdateAll(ctx, tx, updateMap); err != nil {
		t.Error(err)
	} else if rowsAff != 1 {
		t.Error("wanted one record updated but got", rowsAff)
	}
}

func testPaymentStatusesUpsert(t *testing.T) {
	t.Parallel()

	if len(paymentStatusAllColumns) == len(paymentStatusPrimaryKeyColumns) {
		t.Skip("Skipping table with only primary key columns")
	}

	seed := randomize.NewSeed()
	var err error
	// Attempt the INSERT side of an UPSERT
	o := PaymentStatus{}
	if err = randomize.Struct(seed, &o, paymentStatusDBTypes, true); err != nil {
		t.Errorf("Unable to randomize PaymentStatus struct: %s", err)
	}

	ctx := context.Background()
	tx := MustTx(boil.BeginTx(ctx, nil))
	defer func() { _ = tx.Rollback() }()
	if err = o.Upsert(ctx, tx, false, nil, boil.Infer(), boil.Infer()); err != nil {
		t.Errorf("Unable to upsert PaymentStatus: %s", err)
	}

	count, err := PaymentStatuses().Count(ctx, tx)
	if err != nil {
		t.Error(err)
	}
	if count != 1 {
		t.Error("want one record, got:", count)
	}

	// Attempt the UPDATE side of an UPSERT
	if err = randomize.Struct(seed, &o, paymentStatusDBTypes, false, paymentStatusPrimaryKeyColumns...); err != nil {
		t.Errorf("Unable to randomize PaymentStatus struct: %s", err)
	}

	if err = o.Upsert(ctx, tx, true, nil, boil.Infer(), boil.Infer()); err != nil {
		t.Errorf("Unable to upsert PaymentStatus: %s", err)
	}

	count, err = PaymentStatuses().Count(ctx, tx)
	if err != nil {
		t.Error(err)
	}
	if count != 1 {
		t.Error("want one record, got:", count)
	}
}
