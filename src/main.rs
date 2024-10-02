use application_command::ApplicationCommandInteraction;
use command::CommandOptionType;
use dotenv::dotenv;
use serenity::{
    async_trait,
    builder::{CreateApplicationCommand, CreateEmbed},
    model::{
        application::interaction::{Interaction, InteractionResponseType},
        gateway::Ready,
        id::GuildId,
        prelude::*,
    },
    prelude::*,
};
use std::env;

const ROBUX_TO_GBP_RATE: f64 = 0.0035;
const GBP_TO_USD_RATE: f64 = 1.38;
const ROBUX_MARKUP_RATE: f64 = 0.3;

struct Handler;

#[async_trait]
impl EventHandler for Handler {
    async fn interaction_create(&self, ctx: Context, interaction: Interaction) {
        if let Interaction::ApplicationCommand(command) = interaction {
            let result = match command.data.name.as_str() {
                "price" => handle_price_command(&ctx, &command).await,
                "convert" => handle_convert_command(&ctx, &command).await,
                "robux" => handle_robux_command(&ctx, &command).await,
                "help" => handle_help_command(&ctx, &command).await,
                _ => Err(format!("Unknown command: {}", command.data.name)),
            };

            if let Err(error) = result {
                eprintln!("Error handling command: {}", error);
                respond_with_error(&ctx, &command, &error).await;
            }
        }
    }

    async fn ready(&self, ctx: Context, ready: Ready) {
        println!("{} is connected!", ready.user.name);
        if let Err(error) = register_commands(&ctx).await {
            eprintln!("Error registering commands: {}", error);
        }
    }
}

#[tokio::main]
async fn main() -> Result<(), Box<dyn std::error::Error>> {
    dotenv().ok();
    let token = env::var("DISCORD_TOKEN")?;
    let intents = GatewayIntents::GUILD_MESSAGES | GatewayIntents::MESSAGE_CONTENT;

    let mut client = Client::builder(&token, intents)
        .event_handler(Handler)
        .await?;

    client.start().await?;
    Ok(())
}

async fn handle_price_command(
    ctx: &Context,
    command: &ApplicationCommandInteraction,
) -> Result<(), String> {
    let options = &command.data.options;

    if options.len() < 2 {
        return Err("Insufficient command options".to_string());
    }

    let price_type = options[0]
        .value
        .as_ref()
        .ok_or("Missing price type")?
        .as_str()
        .ok_or("Invalid price type")?;
    let amount = options[1]
        .value
        .as_ref()
        .ok_or("Missing amount")?
        .as_u64()
        .ok_or("Invalid amount")? as f64;

    let (rate, is_after_tax) = match price_type {
        "b/t" => (ROBUX_TO_GBP_RATE, false),
        "a/t" => (ROBUX_TO_GBP_RATE / (1.0 - ROBUX_MARKUP_RATE), true),
        _ => return Err("Invalid type. Use 'b/t' or 'a/t'.".to_string()),
    };

    let gbp_amount = amount * rate;
    let gamepass_price = if is_after_tax {
        (amount / (1.0 - ROBUX_MARKUP_RATE)).round() as i64
    } else {
        amount as i64
    };

    let embed = CreateEmbed::default()
        .title("Price Calculation")
        .description(format!(
            "**Conversion Type:** {}\n**Amount of Robux:** {}",
            price_type, amount as i64
        ))
        .field("Gamepass Price", format!("{} R$", gamepass_price), true)
        .field("Amount in GBP", format!("£{:.2}", gbp_amount), true)
        .field(
            "Amount in USD",
            format!("${:.2}", gbp_amount * GBP_TO_USD_RATE),
            true,
        )
        .color(0x0096FF)
        .clone();

    send_embed_response(ctx, command, embed).await
}

async fn handle_convert_command(
    ctx: &Context,
    command: &ApplicationCommandInteraction,
) -> Result<(), String> {
    let options = &command.data.options;

    if options.len() < 2 {
        return Err("Insufficient command options".to_string());
    }

    let currency = options[0]
        .value
        .as_ref()
        .ok_or("Missing currency")?
        .as_str()
        .ok_or("Invalid currency")?;
    let amount = options[1]
        .value
        .as_ref()
        .ok_or("Missing amount")?
        .as_f64()
        .ok_or("Invalid amount")?;

    let (from_currency, to_currency, converted_amount) = match currency {
        "GBP" => ("GBP", "USD", amount * GBP_TO_USD_RATE),
        "USD" => ("USD", "GBP", amount / GBP_TO_USD_RATE),
        _ => return Err("Invalid currency. Use 'GBP' or 'USD'.".to_string()),
    };

    let embed = CreateEmbed::default()
        .title("Currency Conversion")
        .field(
            format!("Amount in {}", from_currency),
            format!("{:.2}", amount),
            true,
        )
        .field(
            format!("Amount in {}", to_currency),
            format!("{:.2}", converted_amount),
            true,
        )
        .color(0x0096FF)
        .clone();

    send_embed_response(ctx, command, embed).await
}

async fn handle_robux_command(
    ctx: &Context,
    command: &ApplicationCommandInteraction,
) -> Result<(), String> {
    let options = &command.data.options;

    if options.len() < 2 {
        return Err("Insufficient command options".to_string());
    }

    let currency = options[0]
        .value
        .as_ref()
        .ok_or("Missing currency")?
        .as_str()
        .ok_or("Invalid currency")?;
    let amount = options[1]
        .value
        .as_ref()
        .ok_or("Missing amount")?
        .as_f64()
        .ok_or("Invalid amount")?;

    let (gbp_amount, usd_amount) = match currency {
        "GBP" => (amount, amount * GBP_TO_USD_RATE),
        "USD" => (amount / GBP_TO_USD_RATE, amount),
        _ => return Err("Invalid currency. Use 'GBP' or 'USD'.".to_string()),
    };

    let robux_amount = (gbp_amount / ROBUX_TO_GBP_RATE) as i64;

    let embed = CreateEmbed::default()
        .title("Robux Calculation")
        .description(format!(
            "{:.2} {} affords {} R$ (£{:.2} / ${:.2})",
            amount, currency, robux_amount, gbp_amount, usd_amount
        ))
        .color(0x0096FF)
        .clone();

    send_embed_response(ctx, command, embed).await
}

async fn handle_help_command(
    ctx: &Context,
    command: &ApplicationCommandInteraction,
) -> Result<(), String> {
    let embed = CreateEmbed::default()
        .title("Available Commands")
        .description(
            "Here are the available commands and their usage:\n\
        /price: Calculate the price in GBP and USD for a given amount of Robux\n\
        /convert: Convert between GBP and USD\n\
        /robux: Convert GBP or USD to the amount of Robux",
        )
        .color(0x0096FF)
        .clone();

    send_embed_response(ctx, command, embed).await
}

async fn send_embed_response(
    ctx: &Context,
    command: &ApplicationCommandInteraction,
    embed: CreateEmbed,
) -> Result<(), String> {
    command
        .create_interaction_response(&ctx.http, |response| {
            response
                .kind(InteractionResponseType::ChannelMessageWithSource)
                .interaction_response_data(|message| message.add_embed(embed))
        })
        .await
        .map_err(|e| format!("Error sending response: {:?}", e))
}

async fn respond_with_error(
    ctx: &Context,
    command: &ApplicationCommandInteraction,
    error_message: &str,
) {
    if let Err(why) = command
        .create_interaction_response(&ctx.http, |response| {
            response
                .kind(InteractionResponseType::ChannelMessageWithSource)
                .interaction_response_data(|message| message.content(error_message))
        })
        .await
    {
        eprintln!("Cannot respond to slash command: {}", why);
    }
}

async fn register_commands(ctx: &Context) -> Result<(), Box<dyn std::error::Error>> {
    let guild_id = GuildId(env::var("GUILD_ID")?.parse()?);

    let commands = guild_id
        .set_application_commands(&ctx.http, |commands| {
            commands
                .create_application_command(|command: &mut CreateApplicationCommand| {
                    command
                        .name("help")
                        .description("Display the available commands and their usage")
                })
                .create_application_command(|command: &mut CreateApplicationCommand| {
                    command
                        .name("price")
                        .description(
                            "Calculate the price in GBP and USD for a given amount of Robux",
                        )
                        .create_option(|option| {
                            option
                                .name("type")
                                .description("Conversion type (b/t or a/t)")
                                .kind(CommandOptionType::String)
                                .required(true)
                                .add_string_choice("b/t", "b/t")
                                .add_string_choice("a/t", "a/t")
                        })
                        .create_option(|option| {
                            option
                                .name("amount")
                                .description("Amount of Robux")
                                .kind(CommandOptionType::Integer)
                                .required(true)
                        })
                })
                .create_application_command(|command: &mut CreateApplicationCommand| {
                    command
                        .name("convert")
                        .description("Convert between GBP and USD")
                        .create_option(|option| {
                            option
                                .name("currency")
                                .description("Currency to convert from (GBP or USD)")
                                .kind(CommandOptionType::String)
                                .required(true)
                                .add_string_choice("GBP", "GBP")
                                .add_string_choice("USD", "USD")
                        })
                        .create_option(|option| {
                            option
                                .name("amount")
                                .description("Amount to convert")
                                .kind(CommandOptionType::Number)
                                .required(true)
                        })
                })
                .create_application_command(|command: &mut CreateApplicationCommand| {
                    command
                        .name("robux")
                        .description("Convert GBP or USD to the amount of Robux")
                        .create_option(|option| {
                            option
                                .name("currency")
                                .description("Currency to convert from (GBP or USD)")
                                .kind(CommandOptionType::String)
                                .required(true)
                                .add_string_choice("GBP", "GBP")
                                .add_string_choice("USD", "USD")
                        })
                        .create_option(|option| {
                            option
                                .name("amount")
                                .description("Amount to convert")
                                .kind(CommandOptionType::Number)
                                .required(true)
                        })
                })
        })
        .await?;

    println!("Registered the following slash commands: {:#?}", commands);
    Ok(())
}
