#![no_std]

imports!();

const TOKEN_NAME: &[u8] = b"TT";


pub trait Exchange {

    #[endpoint(validateGetters)]
    fn validate_getters(&self) -> SCResult<()> {
        sc_try!(self.validate_dcdt_token_name());
        sc_try!(self.validate_dcdt_token_value(5));
        Ok(())
    }

    fn validate_dcdt_token_name(&self) -> SCResult<()> {
        let token_name: Option<Vec<u8>> = self.get_dcdt_token_name();
        match token_name {
            None => {
                sc_error!("dcdt token required")
            },
            Some(name) => {
                require!(name.as_slice() == TOKEN_NAME, "wrong dcdt token");
                Ok(())
            }
        }
    }

    fn validate_dcdt_token_value(&self, expected_value: u64) -> SCResult<()> {
        let token_value = self.get_dcdt_value_big_uint();
        let expected_value = BigUint::from(expected_value);
        require!(expected_value == token_value, "wrong dcdt value");
        Ok(())
    }

    #[endpoint(validateGettersAfterDCDTTransfer)]
    fn validate_getters_after_dcdt_transfer(&self) -> SCResult<()> {
        Ok(())
    }
}
